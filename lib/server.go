package simpleblog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type (
	webfs interface {
		Read(requestFile string) (io.ReadSeeker, error)
	}

	sectionMux map[string]webfs
)

const (
	domainDir     = "./domains/"
	rootDomainDir = "localhost/"
)

// ErrFileNotFound occurs when a requested file is missing
var ErrFileNotFound = errors.New("File not found")

//configTranslator maps strings to correct fs constructors
var configTranslator = map[string]func(string) (webfs, error){
	"blog":  newBfs,
	"media": newMediafs,
}

func newWebfs(path string) (webfs, error) {
	filepath.Clean(path)

	path = filepath.Join(domainDir, path)

	if _, err := os.Stat(path); err != nil {
		return nil, ErrFileNotFound
	}

	read, err := ioutil.ReadFile(filepath.Join(path, "/type"))

	if err != nil {
		return nil, fmt.Errorf("type file not found: %s", err)
	}

	conf := strings.TrimSuffix(string(read), "\n")
	constructor := configTranslator[conf]

	if constructor == nil {
		return nil, fmt.Errorf("Type %s is not defined", conf)
	}
	return constructor(path)
}

//Maps request to file system and serves content
func (sm sectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Access: %s%s by %s\n", r.Host, r.URL.Path, r.RemoteAddr)
	addr := r.Host

	//If the user is connecting on a non standard port
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	if strings.HasPrefix(addr, "www.") {
		addr = strings.Split(addr, "www.")[1]
	}

	if fs := sm.Lookup(addr + "/"); fs != nil {
		requestedFile := r.URL.Path
		requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))
		content, err := fs.Read(requestedFile)
		if err != nil {
			log.Printf("Error: %s for request %s", err, r.URL.Path)
			if err == ErrFileNotFound {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			http.Error(w, "Internal server error", 500)
			return
		}
		http.ServeContent(w, r, requestedFile, time.Now(), content)
		return
	}

	//Nothing found return 404
	http.NotFoundHandler().ServeHTTP(w, r)
}

func (sm sectionMux) Lookup(host string) webfs {
	if fs := sm[host]; fs != nil {
		return fs
	}
	if fs, err := sm.Parse(host); err == nil {
		return fs
	}
	return nil
}

//Parse adds webfs from directory
func (sm sectionMux) Parse(path string) (webfs, error) {
	newfs, err := newWebfs(path)

	if err != nil {
		return nil, fmt.Errorf("Issue creating webfs at %s: %s", path, err)
	}

	sm[path] = newfs

	return newfs, nil
}

//Setup does a first time initalization of the directories
func Setup() error {
	domainRoot := filepath.Join(domainDir, rootDomainDir)

	dirs := []string{
		defaultSourceDir,
		defaultStaticDir,
	}

	pages := map[string]string{
		filepath.Join(defaultSourceDir, "index.md"): indexMessage,
		"page.tmpl":                                 pageTemplate,
		"dir.tmpl":                                  directoryTemplate,
		"type":                                      typeDefault,
	}

	// create directories
	for _, dir := range dirs {
		full := filepath.Join(domainRoot, dir)
		if err := os.MkdirAll(full, 0755); err != nil {
			return fmt.Errorf("setup: failed to create directory '%s'", full)
		}
	}

	// create files
	for key, val := range pages {
		full := filepath.Join(domainRoot, key)
		f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			return fmt.Errorf("setup: failed to create default '%s'", full)
		}

		if _, err := f.WriteString(val); err != nil {
			return fmt.Errorf("setup: failed to write default '%s'", full)
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

//Serve starts a listener with a given port on the given protocol
//currently supported are fcgi(fastcgi) and http
func Serve(port, proto string) error {
	mux := make(sectionMux)
	switch proto {
	case "http":
		err := start(port, mux)
		if err == nil {
			fmt.Println("Server shutdown gracefully")
		}
		return err
	case "fcgi", "fastcgi":
		l, err := net.Listen("tcp", port)
		if err != nil {
			return fmt.Errorf("Serve: Failed to start FCGI client\n%s", err)
		}
		return fcgi.Serve(l, mux)
	default:
		return errors.New("Serve: Protocol not understood")
	}
}

func start(port string, mux http.Handler) error {
	svr := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	errors := make(chan error, 2)

	go func() {
		errors <- svr.ListenAndServe()
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		errors <- fmt.Errorf("%s", <-c)
		close(c)
	}()

	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	<-errors

	return svr.Shutdown(context.TODO())
}

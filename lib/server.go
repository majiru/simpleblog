package simpleblog

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type webfs interface {
	Read(requestFile string) (io.ReadSeeker, error)
}

type sectionMux map[string]webfs

const (
	domainDir     = "./domains/"
	rootDomainDir = "localhost/"
)

//configTranslator maps strings to correct fs constructors
var configTranslator = map[string]func(string) webfs{
	"blog":  newBfs,
	"media": newMediafs,
}

func newWebfs(path string) (webfs, error) {
	read, err := ioutil.ReadFile(filepath.Join(path, "/type"))
	if err != nil {
		return nil, errors.New("type file not found: " + err.Error())
	}
	conf := strings.TrimSuffix(string(read), "\n")
	constructor := configTranslator[conf]
	if constructor == nil {
		return nil, errors.New("Type " + string(conf) + " is not defined")
	}
	return constructor(path), nil
}

//Maps request to file system and serves content
func (sm sectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Access: " + r.Host + r.URL.Path + " by " + r.RemoteAddr)
	addr := r.Host
	//If the user is connecting on a non standard port
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}

	if fs := sm[addr+"/"]; fs != nil {
		requestedFile := r.URL.Path
		requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))
		content, err := fs.Read(requestedFile)
		if err != nil {
			log.Println("Error: " + err.Error() + " for request " + r.URL.Path)
			if err.Error() == "File not found" {
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

//Parse creates sectionMux from directory
func (sm sectionMux) Parse(rootPath string) error {
	_, dirs, err := readDir(rootPath)
	if err != nil {
		return errors.New("Parse: Could not read domain directory\n" + err.Error())
	}

	for _, d := range dirs {
		if strings.HasPrefix(d, "www.") {
			bareHostName := strings.Split(d, "www.")[1]
			newfs, err := newWebfs(rootPath + d)
			if err != nil {
				return errors.New("Issue creating webfs at " + rootPath + d + " : " + err.Error())
			}
			sm[bareHostName] = newfs
			sm[d] = newfs
			continue
		}
		newfs, err := newWebfs(rootPath + d)
		if err != nil {
			return errors.New("Issue creating webfs at " + rootPath + d + " : " + err.Error())
		}
		sm[d] = newfs
	}
	return nil
}

//Setup does a first time initalization of the directories
func Setup() {
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
			log.Printf("setup: failed to create directory '%s'", full)
		}
	}

	// create files
	// todo: if directory wasn't successfully made, don't try to write file
	for key, val := range pages {
		full := filepath.Join(domainRoot, key)
		f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE, 0755)

		if err != nil {
			log.Printf("setup: failed to create default '%s'", full)

			// don't try to write if file wasn't made
			continue
		}

		if _, err := f.WriteString(val); err != nil {
			log.Printf("setup: failed to write default '%s'", full)
		}

		f.Close()
	}
}

//Serve starts a listener with a given port on the given protocol
//currently supported are fcgi(fastcgi) and http
func Serve(port, proto string) error {
	sm := make(sectionMux)
	err := sm.Parse(domainDir)
	if err != nil {
		return errors.New("Serve: Could not parse sections\n" + err.Error())
	}

	switch proto {
	case "http":
		log.Fatal(http.ListenAndServe(port, sm))
	case "fcgi", "fastcgi":
		l, err := net.Listen("tcp", port)
		if err != nil {
			return errors.New("Serve: Failed to start FCGI client\n" + err.Error())
		}
		log.Fatal(fcgi.Serve(l, sm))
	}
	return errors.New("Serve: Protocol not understood")
}

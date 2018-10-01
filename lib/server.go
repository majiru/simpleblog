package simpleblog

import (
	"errors"
	"io"
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

const domainDir = "./domains/"
const rootDomainDir = "localhost/"
const mediaSubDomain = "media."

func (sm sectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (sm sectionMux) Parse(rootPath string) error {
	_, dirs, err := readDir(rootPath)
	if err != nil {
		return errors.New("Parse: Could not read domain directory\n" + err.Error())
	}

	for _, d := range dirs {
		if strings.HasPrefix(d, "www.") {
			bareHostName := strings.Split(d, "www.")[1]
			sm[bareHostName] = newBfs(rootPath + d)
		}
		if strings.HasPrefix(d, mediaSubDomain) {
			sm[d] = &mediafs{rootPath + d}
			continue
		}
		sm[d] = newBfs(rootPath + d)
	}
	return nil
}

/*Setup does a first time initalization of the directories*/
func Setup() {
	os.Mkdir(domainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultSourceDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultStaticDir, 0755)
}

//Serve starts a listener with a given port on the given protocol
//currently supported are fcgi(fastcgi) and http
func Serve(port, proto string) error {
	port = ":" + port
	sm := make(sectionMux)
	err := sm.Parse(domainDir)
	if err != nil {
		return errors.New("Serve: Could not parse sections\n" + err.Error())
	}

	switch proto {
	case "http":
		log.Fatal(http.ListenAndServe(port, sm))
	case "fastcgi":
		fallthrough
	case "fcgi":
		l, err := net.Listen("tcp", port)
		if err != nil {
			return errors.New("Serve: Failed to start FCGI client\n" + err.Error())
		}
		log.Fatal(fcgi.Serve(l, sm))
	}
	return errors.New("Serve: Protocol not understood")
}

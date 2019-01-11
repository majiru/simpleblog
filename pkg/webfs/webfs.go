package webfs

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"aqwari.net/net/styx"
)

//Webfs defines a simple interface to be used for serving web pages
type Webfs interface {
	Read(requestFile string) (io.ReadSeeker, error)
	Stat(path string) (os.FileInfo, error)
}

const domainDir = "./domains/"

//FsMap maps strings to correct fs constructors
type FsMap = map[string]func(string) Webfs

//NewWebfs creates a new webfs using the ConfigTranslator
func NewWebfs(path string, fsmap FsMap) (Webfs, error) {
	filepath.Clean(path)
	path = filepath.Join(domainDir, path)
	if _, err := os.Stat(path); err != nil {
		return nil, errors.New("File not found")
	}
	read, err := ioutil.ReadFile(filepath.Join(path, "/type"))
	if err != nil {
		return nil, errors.New("type file not found: " + err.Error())
	}
	conf := strings.TrimSuffix(string(read), "\n")
	constructor := fsmap[conf]
	if constructor == nil {
		return nil, errors.New("Type " + string(conf) + " is not defined")
	}
	return constructor(path), nil
}

//Server is a struct to allow Webfs to be used with net/http
type Server struct {
	Wfs Webfs
}

func (fs Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedFile := r.URL.Path
	requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))
	content, err := fs.Wfs.Read(requestedFile)
	if err != nil {
		log.Println("Error: " + err.Error() + " for request " + r.URL.Path)
		if err == os.ErrNotExist {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		http.Error(w, "Internal server error", 500)
		return
	}
	http.ServeContent(w, r, requestedFile, time.Now(), content)
	return
}

func (fs Server) Serve9P(s *styx.Session){
	for s.Next(){
		switch msg := s.Request().(type) {
			case styx.Topen:
				msg.Ropen(fs.Wfs.Read(msg.Path()))
			case styx.Twalk:
				msg.Rwalk(fs.Wfs.Stat(msg.Path()))
			case styx.Tstat:
				msg.Rstat(os.Stat(msg.Path()))
		}
	}
}
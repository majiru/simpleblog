package sectionmux

import (
	"errors"
	"fmt"
	"github.com/majiru/simpleblog/pkg/webfs"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//SectionMux serves as a way to map webfs to sub domains
type SectionMux struct {
	mux   map[string]webfs.Webfs
	fsMap webfs.FsMap
}

const (
	domainDir     = "./domains/"
	rootDomainDir = "localhost/"
	indexMessage  = "# Hello from Simpleblog space\n\nThis is your home page.\n"
	typeDefault   = "blog\n"
)

//NewSectionMux initializes a new SectionMux
func NewSectionMux(fsmap webfs.FsMap) SectionMux {
	mux := make(map[string]webfs.Webfs)
	return SectionMux{mux, fsmap}
}

//Maps request to file system and serves content
func (sm SectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Access: " + r.Host + r.URL.Path + " by " + r.RemoteAddr)
	addr := r.Host

	//If the user is connecting on a non standard port
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	if strings.HasPrefix(addr, "www.") {
		addr = strings.Split(addr, "www.")[1]
	}
	fmt.Println(addr + "/")
	if fs := sm.Lookup(addr + "/"); fs != nil {
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

//Lookup checks stored list of webfs before attempting to create a new one
func (sm SectionMux) Lookup(host string) webfs.Webfs {
	if fs := sm.mux[host]; fs != nil {
		return fs
	}

	fs, err := sm.Parse(host)
	if err == nil {
		return fs
	}
	log.Println(err)
	return nil
}

//Parse adds webfs from directory
func (sm SectionMux) Parse(path string) (webfs.Webfs, error) {
	newfs, err := webfs.NewWebfs(path, sm.fsMap)
	if err != nil {
		return nil, errors.New("Issue creating webfs at " + path + " : " + err.Error())
	}
	sm.mux[path] = newfs
	return newfs, nil
}

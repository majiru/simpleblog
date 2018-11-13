package simpleblog

import (
	"errors"
	"github.com/majiru/simpleblog/pkg/blogfs"
	"github.com/majiru/simpleblog/pkg/mediafs"
	"github.com/majiru/simpleblog/pkg/sectionmux"
	"github.com/majiru/simpleblog/pkg/webfs"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
)

const (
	domainDir        = "./domains/"
	rootDomainDir    = "localhost/"
	defaultSourceDir = "source"
	defaultStaticDir = "static"
	templateName     = "page.tmpl"
)

var fsmap = webfs.FsMap{
	"blog":  blogfs.NewBfs,
	"media": mediafs.NewMediafs,
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
	sm := sectionmux.NewSectionMux(fsmap)
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

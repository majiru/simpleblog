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
	err := os.MkdirAll(domainDir+rootDomainDir+defaultSourceDir, 0755)
	if err != nil {
		log.Print("Setup: Failed to create directories, ", err)
	}

	f, err := os.OpenFile(domainDir+rootDomainDir+defaultSourceDir+"/index.md", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Print("Setup: Failed to create default index.md, ", err)
	}
	_, err = f.Write([]byte("# Hello from Simpleblog space\n\nThis is your home page.\n"))
	if err != nil {
		log.Print("Setup: Failed to write default index.md, ", err)
	}
	f.Close()

	f, err = os.OpenFile(domainDir+rootDomainDir+"/page.tmpl", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Print("Setup: Failed to create default page.tmpl, ", err)
	}
	_, err = f.Write([]byte(pageTemplate))
	if err != nil {
		log.Print("Setup: Failed to write default page.tmpl, ", err)
	}
	f.Close()

	if err != nil {
		log.Print("Setup: Failed to initialize default index.md, ", err)
	}

	err = os.Mkdir(domainDir+rootDomainDir+defaultStaticDir, 0755)
	if err != nil {
		log.Print("Setup: Failed to create directory, ", err)
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

const pageTemplate = `
<!DOCTYPE html>
<html>
    <head>
	<meta charset="utf-8">
	<link rel="stylesheet" href="https://unpkg.com/tachyons@4.10.0/css/tachyons.min.css"/>
	<title>{{.Title}}</title>
    </head>
    <body class="bg-washed-yellow pa4">
	<div class="flex flex-wrap justify-around">
	    <div class="w-40 mw5 bg-washed-green bw2 ba pa2 ma3 h-25">
		<ul class="list">
		    {{range $key, $element := .Sidebar}}
		    <div>
			<h3 class="f4 measure-narrow"><a href="{{$key}}">{{$key}}</a></h3>
			<ul>
			{{range $element}}
			    <li class="f5 measure-narrow"><a href="{{.Path}}">{{.Title}}</a></li>
			{{end}}
			</ul>
		    </div>
		    {{end}}
		</ul>
	    </div>
	    <div class="w-80 ba bw2 pa2 ma3 bg-washed-green">
		<h3 class="f1 measure">{{.Title}}</h3>
		{{.Body}}
	    </div>
	</div>
    </body>
</html>
`

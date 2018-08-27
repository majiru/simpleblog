package simpleblog

import (
	"errors"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const defaultSourceDir = "source"
const defaultStaticDir = "static"

type page struct {
	Title   string
	Path    string
	Body    string
	Sidebar map[string][]page
}

type blogfs struct {
	sourceDir string
	staticDir string
}

func (bfs blogfs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedFile := r.URL.Path
	requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))

	if fd, err := os.Stat(bfs.sourceDir + requestedFile); err == nil {
		if fd.IsDir() {
			requestedFile = filepath.Join(requestedFile, "/index.html")
		}
		_, shortName := filepath.Split(requestedFile)
		p, _ := newPage(shortName, requestedFile)
		bfs.getSiblings(p)
		p.read(bfs.sourceDir)
		p.write(w)
		return
	}

	//Check to see if the file exists in the static directory
	if _, err := os.Stat(bfs.staticDir + requestedFile); err == nil {
		http.ServeFile(w, r, bfs.staticDir+requestedFile)
		return
	}

	//Nothing found return 404
	http.NotFoundHandler().ServeHTTP(w, r)
}

func (bfs *blogfs) openDir(path string) (pages, dirpages []page) {
	files, dirs, err := readDir(bfs.sourceDir + path)
	if err != nil {
		log.Fatal("openDir: " + err.Error())
	}
	for _, f := range files {
		p, _ := newPage(f, path+f)
		pages = append(pages, *p)
	}
	for _, d := range dirs {
		p, _ := newPage(d, path+d)
		dirpages = append(dirpages, *p)
	}
	return
}

func (bfs *blogfs) getSiblings(p *page) {
	var siblings = make(map[string][]page)
	dir, _ := filepath.Split(p.Path)
	dirs := strings.Split(dir, "/")
	dirs = dirs[:len(dirs)-1]
	for i := range dirs {
		tempDir := strings.Join(dirs[:i+1], "/")
		tempDir += "/"
		var subdirs []page
		siblings[tempDir], subdirs = bfs.openDir(tempDir)
		siblings[tempDir] = append(siblings[tempDir], subdirs...)

	}
	p.Sidebar = siblings
}

func (p *page) cleanTitle() {
	p.Title = strings.Replace(p.Title, "_", " ", -1)
	p.Title = strings.Title(strings.Split(p.Title, ".html")[0])
	if p.Title == "Index" {
		p.Title = "Home"
	}
}

func (p *page) read(root string) {
	content, err := ioutil.ReadFile(root + p.Path)
	if err != nil {
		log.Fatal("page read: " + err.Error())
	}
	content = blackfriday.Run(content)
	p.Body = string(content)
}

func (p *page) write(dest io.Writer) {
	t, _ := template.New("page").Parse(basicPage)
	t.Execute(dest, p)
}

func readDir(inputDir string) (files, dirs []string, outErr error) {
	infoFiles, err := ioutil.ReadDir(inputDir)
	if err != nil {
		outErr = errors.New("readDir: Could not read dir\n" + err.Error())
		return
	}
	for _, f := range infoFiles {
		if f.IsDir() {
			dirs = append(dirs, f.Name()+"/")
		} else {
			files = append(files, f.Name())
		}
	}
	return
}

func newPage(args ...string) (p *page, err error) {
	p = &page{}
	switch len(args) {
	case 3:
		p.Body = args[2]
		fallthrough
	case 2:
		p.Path = args[1]
		fallthrough
	case 1:
		p.Title = args[0]
	default:
		err = errors.New("newPage: expected 1-3 arguments")
		return
	}
	p.cleanTitle()
	return
}

func newBfs(path string) (bfs *blogfs) {
	bfs = &blogfs{}
	os.Mkdir(path+defaultSourceDir, 0755)
	os.Mkdir(path+defaultStaticDir, 0755)
	bfs.sourceDir = path + defaultSourceDir
	bfs.staticDir = path + defaultStaticDir
	return
}

const basicPage = `
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

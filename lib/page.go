package simpleblog

import (
	"errors"
	"gopkg.in/russross/blackfriday.v2"
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
const defaultBuldDir = "build"

type page struct {
	Title   string
	Path    string
	Body    string
	Sidebar map[string][]page
}

type blogfs struct {
	sourceDir string
	buildDir  string
}

func (bfs blogfs) Open(name string) (http.File, error) {
	fullName := filepath.Join("/", filepath.FromSlash(path.Clean("/"+name)))
	dir, shortName := filepath.Split(fullName)
	p, _ := newPage(shortName, fullName)
	if bfs.needsUpdate(p) {
		bfs.updateStatic(dir)
	}
	f, err := os.Open(bfs.buildDir + fullName)
	if err != nil {
		return nil, errors.New("pageDir Open: Can not open file at " + name)
	}
	return f, nil
}

func (bfs *blogfs) needsUpdate(p *page) bool {
	sourceFile, err := os.Stat(bfs.sourceDir + p.Path)
	if err != nil {
		return false
	}

	destinationFile, err := os.Stat(bfs.buildDir + p.Path)
	if err != nil {
		return true
	}

	if destinationFile.ModTime().Before(sourceFile.ModTime()) {
		return true
	}

	dir, _ := filepath.Split(p.Path)
	dirs, err := ioutil.ReadDir(bfs.buildDir + dir)
	if err != nil {
		log.Fatal("needsUpdate:" + err.Error())
	}

	for _, f := range dirs {
		if !f.IsDir() {
			if f.ModTime().Before(sourceFile.ModTime()) {
				return true
			}
		}
	}
	return false
}

func (bfs *blogfs) updateStatic(path string) {
	pages, dirs := bfs.openDir(path, true)

	for _, p := range pages {
		bfs.getSiblings(&p)
		p.write(bfs.buildDir)
	}
	for _, d := range dirs {
		bfs.updateStatic(d.Path)
	}
}

func (bfs *blogfs) openDir(path string, readBody bool) (pages, dirpages []page) {
	files, dirs, err := readDir(bfs.sourceDir + path)
	if err != nil {
		log.Fatal("openDir: " + err.Error())
	}
	for _, f := range files {
		p, _ := newPage(f, path+f)
		if readBody {
			content, err := ioutil.ReadFile(bfs.sourceDir + path + f)
			if err != nil {
				log.Fatal("openDir: " + err.Error())
			}
			content = blackfriday.Run(content)
			p.Body = string(content)
		}
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
		siblings[tempDir], subdirs = bfs.openDir(tempDir, false)
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

func (p *page) write(root string) {
	fd, err := os.Create(root + p.Path)

	if err != nil {
		dir, _ := filepath.Split(root + p.Path)
		os.Mkdir(dir, 0755)
	}

	t, err := template.New("page").Parse(basicPage)
	t.Execute(fd, p)
}

func readDir(inputDir string) (files, dirs []string, outErr error) {
	infoFiles, err := ioutil.ReadDir(inputDir)
	if err != nil {
		outErr = err
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
	os.Mkdir(path+defaultBuldDir, 0755)
	bfs.sourceDir = path + defaultSourceDir
	bfs.buildDir = path + defaultBuldDir
	return
}

func newBfsFromDir(path string) (bfs map[string]blogfs) {
	bfs = make(map[string]blogfs)
	_, dirs, err := readDir(path)
	if err != nil {
		log.Fatal("newBfsFromDir: " + err.Error())
	}
	for _, d := range dirs {
		bfs[d] = *newBfs(domainDir + d)
	}
	return
}

const basicPage = `<!DOCTYPE html>
<html>
    <head>
	<meta charset="utf-8">
        <link rel="stylesheet" href="/index.css">
        <title>{{.Title}}</title>
        <div class="main">
            <h1>{{.Title}}</h1>
        </div>
    </head>
    <body>
    <div class="sidebar">
    {{range $key, $element := .Sidebar}}
	<h5><a href="{{$key}}">{{$key}}</a></h5>
	<ul>
	{{range $element}}
	    <li><a href="{{.Path}}">{{.Title}}</a></li>
	{{end}}
	</ul>
    {{end}}
    </div>
        <div id="main" class="main">
            {{.Body}}
        </div>
    </body>
</html>
`

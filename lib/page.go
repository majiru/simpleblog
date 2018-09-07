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
const templateName = "page.tmpl"
const defaultTemplate = domainDir + templateName

type page struct {
	Title   string
	Path    string
	Body    string
	Sidebar map[string][]page
}

type blogfs struct {
	sourceDir    string
	staticDir    string
	templateFile string
}

func (bfs blogfs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedFile := r.URL.Path
	requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))

	tmpl, tmplerr := ioutil.ReadFile(bfs.templateFile)
	if tmplerr != nil {
		http.Error(w, "Error finding site template", 500)
		return
	}

	if fd, err := os.Stat(bfs.sourceDir + requestedFile); err == nil {
		if fd.IsDir() {
			requestedFile = filepath.Join(requestedFile, "/index.html")
		}
		_, shortName := filepath.Split(requestedFile)
		p, _ := newPage(shortName, requestedFile)
		bfs.getSiblings(p)
		p.read(bfs.sourceDir)
		if err := p.write(w, string(tmpl)); err != nil {
			http.Error(w, "Erorr parsing template", 500)
		}
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

func (p *page) write(dest io.Writer, tmpl string) error {
	t, err := template.New("page").Parse(tmpl)
	t.Execute(dest, p)
	return err
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

func newPage(args ...string) (*page, error) {
	p := &page{}
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
		return nil, errors.New("newPage: expected 1-3 arguments")
	}
	p.cleanTitle()
	return p, nil
}

func newBfs(path string) *blogfs {
	bfs := &blogfs{path + defaultSourceDir, path + defaultStaticDir, defaultTemplate}
	if fd, err := os.Stat(path + templateName); err == nil {
		bfs.templateFile = path + fd.Name()
	}
	os.Mkdir(path+defaultSourceDir, 0755)
	os.Mkdir(path+defaultStaticDir, 0755)
	return bfs
}

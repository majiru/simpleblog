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

type parser func(in []byte) (out string, err error)

type blogfs struct {
	sourceDir    string
	staticDir    string
	templateFile string
	Parse        parser
}

func (bfs blogfs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestedFile := r.URL.Path
	requestedFile = filepath.Join("/", filepath.FromSlash(path.Clean("/"+requestedFile)))

	//Check to see if the file exists in the static directory
	if _, err := os.Stat(bfs.staticDir + requestedFile); err == nil {
		http.ServeFile(w, r, bfs.staticDir+requestedFile)
		return
	}

	if fd, err := os.Stat(bfs.sourceDir + requestedFile); err == nil {
		if fd.IsDir() {
			requestedFile = filepath.Join(requestedFile, "/index.html")
		}
		_, shortName := filepath.Split(requestedFile)
		p, _ := newPage(shortName, requestedFile)
		bfs.getSiblings(p)
		bfs.read(p)
		if err := bfs.write(w, p); err != nil {
			http.Error(w, "Erorr parsing template "+err.Error(), 500)
		}
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

func (bfs *blogfs) read(p *page) error {
	content, err := ioutil.ReadFile(bfs.sourceDir + p.Path)
	if err != nil {
		return errors.New("bfs.Read: " + err.Error())
	}
	body, err := bfs.Parse(content)
	if err != nil {
		return errors.New("bfs.Read: " + err.Error())
	}
	p.Body = body
	return nil
}

func (bfs *blogfs) write(dest io.Writer, p *page) error {
	t, err := template.ParseFiles(bfs.templateFile)
	if err != nil {
		return errors.New("bfs.Write: " + err.Error())
	}

	err = t.Execute(dest, p)
	if err != nil {
		return errors.New("bfs.Write: " + err.Error())
	}
	return err
}

func markdown2html(in []byte) (string, error) {
	return string(blackfriday.Run(in)), nil
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

func (p *page) cleanTitle() {
	p.Title = strings.Replace(p.Title, "_", " ", -1)
	p.Title = strings.Title(strings.Split(p.Title, ".html")[0])
	if p.Title == "Index" {
		p.Title = "Home"
	}
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

func newBfs(path string, p parser) *blogfs {
	if p == nil {
		p = markdown2html
	}
	bfs := &blogfs{path + defaultSourceDir, path + defaultStaticDir, defaultTemplate, markdown2html}
	if fd, err := os.Stat(path + templateName); err == nil {
		bfs.templateFile = path + fd.Name()
	}
	os.Mkdir(path+defaultSourceDir, 0755)
	os.Mkdir(path+defaultStaticDir, 0755)
	return bfs
}

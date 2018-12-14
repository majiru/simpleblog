package blogfs

import (
	"bytes"
	"errors"
	"github.com/majiru/simpleblog/pkg/basicfs"
	"github.com/majiru/simpleblog/pkg/page"
	"github.com/majiru/simpleblog/pkg/webfs"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const defaultSourceDir = "source"
const defaultStaticDir = "static"
const templateName = "page.tmpl"

type blogfs struct {
	sourceDir    string
	staticDir    string
	templateFile string
}

//NewBfs creates a new blogfs webfs
func NewBfs(path string) webfs.Webfs {
	source := filepath.Join(path, defaultSourceDir)
	static := filepath.Join(path, defaultStaticDir)
	templ := filepath.Join(path, templateName)
	globalTempl := filepath.Join(path, "..", templateName)
	bfs := &blogfs{source, static, globalTempl}
	if _, err := os.Stat(templ); err == nil {
		bfs.templateFile = templ
	}
	os.Mkdir(path+defaultSourceDir, 0755)
	os.Mkdir(path+defaultStaticDir, 0755)
	return bfs
}

func (bfs *blogfs) Read(request string) (io.ReadSeeker, error) {
	if fd, err := os.Stat(bfs.sourceDir + request); err == nil {
		if fd.IsDir() {
			request = filepath.Join(request, "/index.md")
		}
		content, _ := ioutil.ReadFile(bfs.sourceDir + request)
		content = blackfriday.Run(content)

		_, shortName := filepath.Split(request)
		p, _ := page.NewPage(shortName, request, string(blackfriday.Run(content)))
		bfs.getSiblings(p)
		t, err := template.ParseFiles(bfs.templateFile)
		if err != nil {
			return nil, errors.New("Template file not found")

		}
		var out bytes.Buffer
		if err := t.Execute(&out, p); err != nil {
			return nil, errors.New("Error processing template")
		}

		return strings.NewReader(out.String()), nil

	}

	if fd, err := os.Open(bfs.staticDir + request); err == nil {
		return fd, nil
	}
	return nil, errors.New("File not found")
}

func (bfs *blogfs) openDir(path string) (pages, dirpages []page.Page, err error) {
	files, dirs, err := basicfs.List(bfs.sourceDir + path)
	if err != nil {
		return
	}
	for _, f := range files {
		p, _ := page.NewPage(f, path+f)
		pages = append(pages, *p)
	}
	for _, d := range dirs {
		p, _ := page.NewPage(d, path+d)
		dirpages = append(dirpages, *p)
	}
	return
}

func (bfs *blogfs) getSiblings(p *page.Page) error {
	var siblings = make(map[string][]page.Page)
	dir, _ := filepath.Split(p.Path)
	dirs := strings.Split(dir, "/")
	dirs = dirs[:len(dirs)-1]
	for i := range dirs {
		tempDir := strings.Join(dirs[:i+1], "/")
		tempDir += "/"
		files, subdirs, err := bfs.openDir(tempDir)
		if err != nil {
			return errors.New("Error parsing dir" + tempDir)
		}
		siblings[tempDir] = append(files, subdirs...)

	}
	p.Sidebar = siblings
	return nil
}

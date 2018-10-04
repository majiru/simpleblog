package simpleblog

import (
	"bytes"
	"errors"
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
const defaultTemplate = domainDir + templateName

type blogfs struct {
	sourceDir    string
	staticDir    string
	templateFile string
}

func newBfs(path string) webfs {
	bfs := &blogfs{path + defaultSourceDir, path + defaultStaticDir, defaultTemplate}
	if fd, err := os.Stat(path + templateName); err == nil {
		bfs.templateFile = path + fd.Name()
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
		p, _ := newPage(shortName, request, string(blackfriday.Run(content)))
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

func (bfs *blogfs) openDir(path string) (pages, dirpages []page, err error) {
	files, dirs, err := readDir(bfs.sourceDir + path)
	if err != nil {
		return
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

func (bfs *blogfs) getSiblings(p *page) error {
	var siblings = make(map[string][]page)
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

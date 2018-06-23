package simpleblog

import (
	"errors"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const buildDir = "./build"
const sourceDir = "./source"

type page struct {
	Title      string
	Outputfile string
	Body       string
}

func (p *page) cleanTitle() {
	p.Title = strings.Replace(p.Title, "_", " ", -1)
	p.Title = strings.Title(strings.Split(p.Title, ".html")[0])
}

func (p page) GetHeader() map[string][]page {
	var output = make(map[string][]page)
	dir, _ := filepath.Split(p.Outputfile)
	dirs := strings.Split(dir, "/")
	dirs = dirs[:len(dirs)-1]
	for i := range dirs {
		tempDir := strings.Join(dirs[:i+1], "/")
		tempDir += "/"
		var subdirs []string
		output[tempDir], subdirs = newPagesFromDir(tempDir)
		for _, subdir := range subdirs {
			p, _ := newPage(subdir, tempDir+subdir)
			output[tempDir] = append(output[tempDir], p)
		}
	}
	return output
}

func (p page) write() {
	fd, err := os.Create(buildDir + p.Outputfile)

	if err != nil {
		dir, _ := filepath.Split(buildDir + p.Outputfile)
		os.Mkdir(dir, 0755)
	}

	t, err := template.New("page").Parse(basicPage)
	t.Execute(fd, p)
}

/*
  updatePath updates all of the content files recursivly down tree path
*/
func updatePath(path string) {
	pages, dirs := newPagesFromDir(path)
	for _, p := range pages {
		content, _ := ioutil.ReadFile(sourceDir + p.Outputfile)
		content = blackfriday.Run(content)
		p.Body = string(content)
		p.write()
	}
	for _, dir := range dirs {
		updatePath(path + dir)
	}
}

func newPagesFromDir(path string) ([]page, []string) {
	var pages []page
	var dirs []string
	files, _ := ioutil.ReadDir(sourceDir + path)
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f.Name()+"/")
		} else {
			p, _ := newPage(f.Name(), path+f.Name())
			pages = append(pages, p)
		}
	}
	return pages, dirs
}

func newPage(args ...string) (page, error) {
	p := page{}
	switch len(args) {
	case 3:
		p.Body = args[2]
		fallthrough
	case 2:
		p.Outputfile = args[1]
		fallthrough
	case 1:
		p.Title = args[0]
	default:
		return page{}, errors.New("newPage: Error: expected 1-3 arguments")

	}
	p.cleanTitle()
	return p, nil
}

const basicPage = `<!DOCTYPE html>
<html>
    <head>
        <link rel="stylesheet" href="/index.css">
        <title>{{.Title}}</title>
        <div class="main">
            <h1>{{.Title}}</h1>
        </div>
    </head>
    <body>
    <div class="sidebar">
    {{range $key, $element := .GetHeader}}
	<h5><a href="{{$key}}">{{$key}}</a></h5>
	<ul>
	{{range $element}}
	    <li><a href="{{.Outputfile}}">{{.Title}}</a></li>
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

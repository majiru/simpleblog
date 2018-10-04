package simpleblog

import (
	"errors"
	"io/ioutil"
	"strings"
)

type page struct {
	Title   string
	Path    string
	Body    string
	Sidebar map[string][]page
}

func (p *page) cleanTitle() {
	p.Title = strings.Replace(p.Title, "_", " ", -1)
	p.Title = strings.Title(strings.Split(p.Title, ".md")[0])
	if p.Title == "Index" {
		p.Title = "Home"
	}
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

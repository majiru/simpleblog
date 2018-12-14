package page

import (
	"errors"
	"strings"
)

//Page acts as way to organize data before being sent to a template
type Page struct {
	Title   string
	Path    string
	Body    string
	Sidebar map[string][]Page
}

func (p *Page) cleanTitle() {
	p.Title = strings.Replace(p.Title, "_", " ", -1)
	p.Title = strings.Title(strings.Split(p.Title, ".md")[0])
	if p.Title == "Index" {
		p.Title = "Home"
	}
}

//NewPage creates a new Page struct
func NewPage(args ...string) (*Page, error) {
	p := &Page{}
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

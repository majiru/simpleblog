package simpleblog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type mediafs struct {
	root string
}

func (mfs *mediafs) Read(request string) (io.ReadSeeker, error) {
	if fi, err := os.Stat(mfs.root + request); err == nil {
		if !fi.IsDir() {
			if fd, err := os.Open(mfs.root + request); err == nil {
				return fd, err
			}
		}
		out, err := mfs.openDir(request)
		if err != nil {
			return strings.NewReader(err.Error()), nil
		}
		return out, nil

	}
	return nil, errors.New("File not found")
}

func (mfs *mediafs) openDir(path string) (io.ReadSeeker, error) {
	files, dirs, err := readDir(mfs.root + path)
	if err != nil {
		return nil, err
	}
	p, _ := newPage("File Browser", "/")
	directory := make(map[string][]page)
	directory["root"] = []page{}
	for _, f := range files {
		filepath.Join(path, "/")
		listing, _ := newPage(f, path+f)
		directory["root"] = append(directory["root"], *listing)
	}
	for _, d := range dirs {
		listing, _ := newPage(d, path+d)
		directory["root"] = append(directory["root"], *listing)
	}
	p.Sidebar = directory
	var out bytes.Buffer
	t, err := template.New("directory").Parse(directoryTemplate)
	if err != nil {
		return nil, errors.New("Template not found")
	}
	if err := t.Execute(&out, p); err != nil {
		return nil, errors.New("Error processing template")
	}

	return strings.NewReader(out.String()), nil
}

const directoryTemplate = `
<!DOCTYPE html>
<html>
    <head>
	<meta charset="utf-8">
	<link rel="stylesheet" href="https://unpkg.com/tachyons@4.10.0/css/tachyons.min.css"/>
	<title>{{.Title}}</title>
    </head>
    <body class="bg-washed-yellow pa4">
	<div class="ba4 bw2 pa2 ma3 bg-washed-green">
	    <h3 class="f1 measure">{{.Title}}</h3>
	    <ul>
		 {{range $key, $element := .Sidebar}}
		    {{range $element}}
			<li class="f5 measure-narrow"><a href="{{.Path}}">{{.Title}}</a></li>
		    {{end}}
		{{end}}
	    </ul>
	</div>
    </body>
</html>
`

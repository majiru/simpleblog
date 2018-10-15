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

const dirTemplName = "dir.tmpl"

var defaultDirTempl = filepath.Join(domainDir, dirTemplName)

type mediafs struct {
	root string
}

func newMediafs(root string) (webfs, error) {
	contentDir := filepath.Join(root, "media")

	if err := os.Mkdir(contentDir, 0644); err != nil {
		return nil, err
	}

	return &mediafs{contentDir}, nil
}

func (mfs *mediafs) Read(request string) (io.ReadSeeker, error) {
	path := filepath.Join(mfs.root, request)
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			if fd, err := os.Open(path); err == nil {
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
		listing, _ := newPage(f, filepath.Join(path, f))
		directory["root"] = append(directory["root"], *listing)
	}
	for _, d := range dirs {
		listing, _ := newPage(d, filepath.Join(path, d))
		directory["root"] = append(directory["root"], *listing)
	}
	p.Sidebar = directory
	var out bytes.Buffer
	t, err := template.ParseFiles(defaultDirTempl)
	if err != nil {
		return nil, errors.New("Template not found")
	}
	if err := t.Execute(&out, p); err != nil {
		return nil, errors.New("Error processing template")
	}

	return strings.NewReader(out.String()), nil
}

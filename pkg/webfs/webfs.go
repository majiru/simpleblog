package webfs

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//Webfs defines a simple interface to be used for serving web pages
type Webfs interface {
	Read(requestFile string) (io.ReadSeeker, error)
}

const (
	domainDir     = "./domains/"
	rootDomainDir = "localhost/"
	indexMessage  = "# Hello from Simpleblog space\n\nThis is your home page.\n"
	typeDefault   = "blog\n"
)

//FsMap maps strings to correct fs constructors
type FsMap = map[string]func(string) Webfs

//NewWebfs creates a new webfs using the ConfigTranslator
func NewWebfs(path string, fsmap FsMap) (Webfs, error) {
	filepath.Clean(path)
	path = filepath.Join(domainDir, path)
	if _, err := os.Stat(path); err != nil {
		return nil, errors.New("File not found")
	}
	read, err := ioutil.ReadFile(filepath.Join(path, "/type"))
	if err != nil {
		return nil, errors.New("type file not found: " + err.Error())
	}
	conf := strings.TrimSuffix(string(read), "\n")
	constructor := fsmap[conf]
	if constructor == nil {
		return nil, errors.New("Type " + string(conf) + " is not defined")
	}
	return constructor(path), nil
}

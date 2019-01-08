package basicfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

//Walk recursivly calls functions on a directory
func Walk(root string, fileFunc func(*os.File), dirFunc func(*os.File)) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if fd, err := os.Open(filepath.Join(path, info.Name())); err == nil {
				dirFunc(fd)
			}
			return err
		}
		if fd, err := os.Open(filepath.Join(path, info.Name())); err == nil {
			fileFunc(fd)
		}
		return err
	})
}

//List splits the contents of the dir into file and dir string slices
func List(dir string) (files []string, dirs []string, err error) {
	infoFiles, err := ioutil.ReadDir(dir)
	if err != nil {
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

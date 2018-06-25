package simpleblog

import (
	"log"
	"net/http"
	"os"
)

/*Update serves the content with the specific handler*/
func Update() {
	updatePath("/")
}

/*Setup does a first time initalization of the directory*/
func Setup() {
	os.Mkdir(buildDir, 0755)
	os.Mkdir(sourceDir, 0755)
}

/*Serve creates a http server serving the content*/
func Serve() {
	http.Handle("/", http.FileServer(pageDir("/")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package simpleblog

import (
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type sectionMux map[string]http.Handler

const domainDir = "./domains/"
const rootDomainDir = "root/"

func (sm sectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := sm[r.Host]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, r.Host, 403)
	}
}

func (sm sectionMux) Parse(fsIndex map[string]blogfs, hostname string) {
	sm[hostname] = http.FileServer(fsIndex[rootDomainDir])
	sm["www."+hostname] = http.FileServer(fsIndex[rootDomainDir])
	delete(fsIndex, "root")

	for k, v := range fsIndex {
		k = k[:len(k)-1]
		k += "."
		sm[k+hostname] = http.FileServer(v)
	}
}

/*Build Outputes */
func Build() {
	bfs := newBfsFromDir(domainDir)
	for _, fs := range bfs {
		fs.updateStatic("/")
	}
}

/*Setup does a first time initalization of the directories*/
func Setup() {
	os.Mkdir(domainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultSourceDir, 0755)
}

/*Servefcgi starts a FastCGI listener using a sectionMux*/
func Servefcgi(hostname, port string) {
	port = ":" + port
	sm := make(sectionMux)
	sm.Parse(newBfsFromDir(domainDir), hostname)

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fcgi.Serve(l, sm))
}

/*Serve serves the root domain over HTTP*/
func Serve(port string) {
	port = ":" + port
	bfs := newBfs(domainDir + rootDomainDir)
	bfs.updateStatic("/")
	http.Handle("/", http.FileServer(bfs))
	log.Fatal(http.ListenAndServe(port, nil))
}

package simpleblog

import (
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
)

type sectionMux map[string]*blogfs

const domainDir = "./domains/"
const rootDomainDir = "localhost/"

func (sm sectionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := sm.route(r.Host); handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, r.Host, 403)
	}
}

func (sm sectionMux) route(addr string) http.Handler {
	//If the user is connecting on a non 80 port
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	if fs := sm[addr]; fs != nil {
		return http.FileServer(*fs)
	}
	//In the event that the requested page is a directory
	if fs := sm[addr+"/"]; fs != nil {
		return http.FileServer(*fs)
	}
	//Nothing found return 404
	return http.NotFoundHandler()
}

func (sm sectionMux) Parse(rootPath string) {
	_, dirs, err := readDir(rootPath)
	if err != nil {
		log.Fatal("Could not read domain directory")
	}

	for _, d := range dirs {
		if strings.HasPrefix(d, "www.") {
			bareHostName := strings.Split(d, "www.")[1]
			sm[bareHostName] = newBfs(rootPath + d)
		}
		sm[d] = newBfs(rootPath + d)
	}

}

/*Build Outputes */
func Build() {
	sm := make(sectionMux)
	sm.Parse(domainDir)
	for _, bfs := range sm {
		bfs.updateStatic("/")
	}
}

/*Setup does a first time initalization of the directories*/
func Setup() {
	os.Mkdir(domainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultSourceDir, 0755)
}

/*Servefcgi starts a FastCGI listener using a sectionMux*/
func Servefcgi(port string) {
	port = ":" + port
	sm := make(sectionMux)
	sm.Parse(domainDir)

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fcgi.Serve(l, sm))
}

/*Serve serves the root domain over HTTP*/
func Serve(port string) {
	port = ":" + port
	sm := make(sectionMux)
	sm.Parse(domainDir)
	log.Fatal(http.ListenAndServe(port, sm))
}

package simpleblog

import (
	"errors"
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
	addr := r.Host
	//If the user is connecting on a non standard port
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	if fs := sm[addr+"/"]; fs != nil {
		fs.ServeHTTP(w, r)
		return
	}
	//Nothing found return 404
	http.NotFoundHandler().ServeHTTP(w, r)
}

func (sm sectionMux) Parse(rootPath string) error {
	_, dirs, err := readDir(rootPath)
	if err != nil {
		return errors.New("Parse: Could not read domain directory\n" + err.Error())
	}

	for _, d := range dirs {
		if strings.HasPrefix(d, "www.") {
			bareHostName := strings.Split(d, "www.")[1]
			sm[bareHostName] = newBfs(rootPath + d)
		}
		sm[d] = newBfs(rootPath + d)
	}
	return nil
}

/*Setup does a first time initalization of the directories*/
func Setup() {
	os.Mkdir(domainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultSourceDir, 0755)
	os.Mkdir(domainDir+rootDomainDir+defaultStaticDir, 0755)
}

//Serve starts a listener with a given port on the given protocol
//currently supported are fcgi(fastcgi) and http
func Serve(port, proto string) error {
	port = ":" + port
	sm := make(sectionMux)
	err := sm.Parse(domainDir)
	if err != nil {
		return errors.New("Serve: Could not parse sections\n" + err.Error())
	}

	switch proto {
	case "http":
		log.Fatal(http.ListenAndServe(port, sm))
	case "fastcgi":
		fallthrough
	case "fcgi":
		l, err := net.Listen("tcp", port)
		if err != nil {
			return errors.New("Serve: Failed to start FCGI client\n" + err.Error())
		}
		log.Fatal(fcgi.Serve(l, sm))
	}
	return errors.New("Serve: Protocol not understood")
}

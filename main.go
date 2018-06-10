package main

import (
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type page struct {
	Title      string
	Outputfile string
	Body       string
	Section    []int
}

type link struct {
	Title    string
	Filename string
}

const websiteTitle = "Your Title Here"

var (
	rootPath  = "html"
	buildPath = "static/"
	links     [][]link
	linkIndex = 1
)

var funcMap = template.FuncMap{
	"getLinks": getLinks,
}

func init() {
	links = append(links, []link{})
	os.Mkdir(buildPath, 0755)
}

func main() {
	pages := processSection(rootPath, websiteTitle)
	blogs := processSection("blog", "Blogs")

	links[0][0].Title = "Home"

	writeStatic(pages)
	writeStatic(blogs)

	http.Handle("/", http.FileServer(http.Dir(buildPath)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getLinks(index int) []link {
	return links[index]
}

func nameToTitle(filename string) string {
	output := strings.Replace(filename, "_", " ", -1)
	output = strings.Title(strings.Split(output, ".html")[0])
	return output
}

func processSection(filepath, title string) []page {
	files, err := ioutil.ReadDir(filepath)

	if err != nil {
		log.Panic(err)
	}

	var tempLinks []link
	var tempPages []page

	for _, f := range files {
		content, _ := ioutil.ReadFile(filepath + "/" + f.Name())
		content = blackfriday.Run(content)

		newPage := page{title, f.Name(), string(content), []int{0, linkIndex}}

		if filepath != rootPath {
			newPage.Outputfile = filepath + "/" + newPage.Outputfile
			os.Mkdir(buildPath+filepath, 0755)
		}

		newLink := link{nameToTitle(f.Name()), newPage.Outputfile}

		if f.Name() == "index.html" {
			newLink.Title = title
			links[0] = append(links[0], newLink)
		} else {
			tempLinks = append(tempLinks, newLink)
		}
		tempPages = append(tempPages, newPage)
	}
	links = append(links, tempLinks)
	linkIndex++
	return tempPages
}

func writeStatic(pages []page) {
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles("main.tmpl")
	if err != nil {
		log.Panic(err)
	}
	for _, p := range pages {
		outputFile, _ := os.Create(buildPath + p.Outputfile)
		err = tmpl.ExecuteTemplate(outputFile, "main.tmpl", p)
		if err != nil {
			log.Fatal(err)

		}
	}
}

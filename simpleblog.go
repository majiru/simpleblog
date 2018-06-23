package main

import (
	"fmt"
	"github.com/majiru/simpleblog/lib"
	"os"
)

const usage = `
simpleblog {init, build, run}
    init: Creates program dirs
    build: Writes output to files
    run: Writes output to files and serves content
`

func main() {
	needsHelp := false

	if len(os.Args) < 2 {
		needsHelp = true
	}

	for _, arg := range os.Args[1:] {
		switch arg {
		case "init":
			simpleblog.Setup()
		case "build":
			simpleblog.Update()
		case "run":
			simpleblog.Update()
			simpleblog.Serve()
		default:
			needsHelp = true
			fmt.Println("Arg: '" + arg + "' not understood")
		}
	}

	if needsHelp {
		printUsage()
	}

}

func printUsage() {
	fmt.Println(usage)
	os.Exit(1)
}

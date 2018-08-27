package main

import (
	"flag"
	"fmt"
	"github.com/majiru/simpleblog/lib"
)

const usage = `
simpleblog  {init, build, run}
Commands:
    init: Creates program dirs
    build: Writes output to files
    run: Writes output to files and serves content
`

func main() {
	var port = flag.String("Port", "8080", "Port to run service on")
	var protocol = flag.String("Protocol", "http", "http or fcgi")

	flag.StringVar(port, "p", *port, "Port to run service on")
	flag.StringVar(protocol, "r", *protocol, "http or fcgi")
	flag.Parse()

	needsHelp := false

	if len(flag.Args()) < 1 {
		needsHelp = true
	}

	for _, arg := range flag.Args() {
		switch arg {
		case "init":
			simpleblog.Setup()
		case "build":
			simpleblog.Build()
		case "run":
			switch *protocol {
			case "http":
				simpleblog.Serve(*port)
			case "fcgi":
				fallthrough
			case "fastcgi":
				simpleblog.Servefcgi(*port)
			default:
				needsHelp = true
				fmt.Println("Protocol: '" + *protocol + "' not understood")
			}
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
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

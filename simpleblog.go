package main

import (
	"flag"
	"fmt"
	"github.com/majiru/simpleblog/lib"
)

const usage = `
simpleblog  {init, run}
Commands:
    init: Creates program dirs
    run: Serve content
`

func main() {
	var port = flag.String("port", "8080", "Port to run service on")
	var protocol = flag.String("protocol", "http", "http or fcgi")

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
		case "run":
			if err := simpleblog.Serve(*port, *protocol); err != nil {
				needsHelp = true
				fmt.Println(err.Error())
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

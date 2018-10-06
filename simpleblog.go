package main

import (
	"flag"
	"fmt"
	"github.com/majiru/simpleblog/lib"
	"log"
	"os"
)

func main() {
	port := flag.Int("p", 8080, "Port to run service on")
	protocol := flag.String("r", "http", "http or fcgi")

	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		return
	}

	arg := flag.Args()[0]

	switch arg {
	case "init":
		simpleblog.Setup()

		if flag.NArg() < 2 || flag.Args()[1] != "run" {
			break
		}

		fallthrough
	case "run":
		tailport := fmt.Sprintf(":%d", *port)
		if err := simpleblog.Serve(tailport, *protocol); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("arg '%s' not understood", arg)
	}
}

func printUsage() {
	const usage = `simpleblog [arguments] {init, run}

Commands:
    init: Creates program dirs
    run:  Serve content

Flags:`

	fmt.Fprintln(os.Stderr, usage)

	flag.PrintDefaults()

	os.Exit(2)
}

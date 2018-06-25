# Simpleblog

This is a simple web application that translates folders with markdown files in them into a very basic navigatable site. This was desgined to make adding content easy and efficent, without having to dell with any html. 

## Features

* Translates folders and markdown files into navigatable website
* Serves content on a webserver with net/http golang library
* Dynamically processes requests to check for source updates
* Very small, weighing in at just about 200 lines
* Only uses http.Handle(), meaning it never does 'manual' parsing of incoming requests and simply passes the file descriptor to the http server after updating it's content
* Program structure makes it easy to expand on or change if you happen to have other design needs


## Setup

First, download simpleblog using the `go get` command:


`go get github.com/majiru/simpleblog`

## Usage
Navigate to a directory you want to use for website and run `simpleblog init` to create the basic directories.

From there you can create markdown html files and directiores in the ./source directory

After creating your content, simply run `simpleblog build` to have simpleblog build the output files into ./build


From there you can run `simpleblog run` to serve the content on the localport 8080, `run` will also update the content before starting the web server. Simpleblog will also check the source dir each time the user requests a page and update the build directory before serving the file.


There is an example directory included in this repo under ./example/, with sample source and build directories
## Reasoning

I dislike over the top JS and prefer something a bit simpler with the web pages, this allows me to write simple text based pages without having to copy the header in to each file.

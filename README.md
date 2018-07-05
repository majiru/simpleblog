# Simpleblog

This is a simple web application that translates folders with markdown files in them into a very basic navigatable site. This was desgined to make adding content easy and efficent, without having to dell with any html. 

## Features

* Subdomains using FastCGI
* Translates folders and markdown files into navigatable website
* Serves content on a webserver with net/http golang library
* Dynamically processes requests to check for source updates
* Very small, weighing in at under 300 lines
* Program structure makes it easy to expand on or change if you happen to have other design needs


## Setup

First, download simpleblog using the `go get` command:


`go get github.com/majiru/simpleblog`

## Usage
Navigate to a directory you want to use for website and run `simpleblog init` to create the basic directories.

This will create a directory structure like this:
    * domains
        * root
            * source

Create or copy markdown files into the source directory with their build names( index.html, about_thing.html )

After creating or copying content into the source folder, run `simpleblog build` to generate the output files for testing

From there you can run `simpleblog run` to serve the content over http on port 8080, there is a (-port, -p) flag that changes this port.

There is an example directory containing source and build directories with this repo.

## Subdomains
If you would like to create other sub domains, create other directories in the domains directory, for example a folder blog would corelate to blog.example.com. Markdown files should be placed in a similar fashion to the root domain. Note that multiple domains requires the use of f

An example of a directory tree with a blog subdmoain with some content might look like this:
    *domains
        *root
            *source
                *index.html
        *blog
            *source
                *index.html

Build files are stored inside of a build folder under each subdomain folder.

In order to take advantage of subdomains the FastCGI protocol must be used, this can be activated with the (-protocol, -r) flags, most often the hostname will also need to be set so that simpleblog can properly route the connections. This is done with the (-hostname, -h) flags.

As an example `simpleblog -r fcgi -h example.com build run` will build and run a fastcgi server with hostname 'example.com'

## Reasoning

I dislike over the top JS and prefer something a bit simpler with the web pages, this allows me to write simple text based pages without having to copy the header in to each file.

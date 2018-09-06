# Simpleblog

This is a simple web application that translates markdown text files into html pages surronded by a go text template.

## Features

* Subdomain support
* Translates folders and markdown files into navigatable website
* Aimed at being lightweight and simple


## Setup

First, download simpleblog using the `go get` command:

`go get github.com/majiru/simpleblog`

## Usage
Navigate to a directory you want to use for website and run `simpleblog init` to create the basic directories.

This will create a directory structure like this:
* domains
    * localhost
      * source
      * static

Create or copy markdown files into the source directory with their build names( index.html, about_thing.html )

From there you can run `simpleblog run` to serve the content over http on port 8080, there is a (-port, -p) flag that changes this port.

There is an example directory containing source and static directories with this repo.

## Subdomains
If you would like to create other sub domains, simply create a folder of the FQDN in the domains folder. Markdown files should be placed in a similar fashion to the root domain. 



An example of a directory tree with a blog subdmoain with some content might look like this:
* domains
  * localhost
    * static
    * source
      * index.html
  * mywebsite.org
    * static
    * source
      * index.html
  * blog.mywebsite.org
    * static
    * source
      * index.html

Note that if you are using some sort of web server between simpleBlog and the internet then simpleBlog may not be able to gather the original host request which is needed for subdomains. This can be solved by switching the protocol to FastCGI using the (-protocol -r) flags.

## Reasoning

I wanted something like werc and thought it would be fun to build a minimalistic web app.

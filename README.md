# Simpleblog

This is a simple static http server, made to serve simple text based content. I wrote this to create a simple blog system that I can easily drop posts in to.

## Usage

Simply drop markdown files into the html or blog folder as (title).html, each folder will also need a index.html home page.
Simpleblog will parse these files once on boot and then serve them statically to the user, this means the web server does not have to pass through golang if desired and the static output could be copied elsewhere.

## Reasoning

I dislike over the top JS and prefer something a bit simpler with the web pages, this allows me to write simple text based pages without having to copy the header in to each file.

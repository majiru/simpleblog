package simpleblog

const (
	pageTemplate = `
<!DOCTYPE html>
<html>
    <head>
	<meta charset="utf-8">
	<link rel="stylesheet" href="https://unpkg.com/tachyons@4.10.0/css/tachyons.min.css"/>
	<title>{{.Title}}</title>
    </head>
    <body class="bg-washed-yellow pa4">
	<div class="flex flex-wrap justify-around">
	    <div class="w-40 mw5 bg-washed-green bw2 ba pa2 ma3 h-25">
		<ul class="list">
		    {{range $key, $element := .Sidebar}}
		    <div>
			<h3 class="f4 measure-narrow"><a href="{{$key}}">{{$key}}</a></h3>
			<ul>
			{{range $element}}
			    <li class="f5 measure-narrow"><a href="{{.Path}}">{{.Title}}</a></li>
			{{end}}
			</ul>
		    </div>
		    {{end}}
		</ul>
	    </div>
	    <div class="w-80 ba bw2 pa2 ma3 bg-washed-green">
		<h3 class="f1 measure">{{.Title}}</h3>
		{{.Body}}
	    </div>
	</div>
    </body>
</html>
`

	directoryTemplate = `
<!DOCTYPE html>
<html>
    <head>
	<meta charset="utf-8">
	<link rel="stylesheet" href="https://unpkg.com/tachyons@4.10.0/css/tachyons.min.css"/>
	<title>{{.Title}}</title>
    </head>
    <body class="bg-washed-yellow pa4">
	<div class="ba4 bw2 pa2 ma3 bg-washed-green">
	    <h3 class="f1 measure">{{.Title}}</h3>
	    <ul>
		 {{range $key, $element := .Sidebar}}
		    {{range $element}}
			<li class="f5 measure-narrow"><a href="{{.Path}}">{{.Title}}</a></li>
		    {{end}}
		{{end}}
	    </ul>
	</div>
    </body>
</html>
`

	indexMessage = "# Hello from Simpleblog space\n\nThis is your home page.\n"

	typeDefault = "blog\n"
)

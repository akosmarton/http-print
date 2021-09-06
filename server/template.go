package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Template struct {
	tpl *template.Template
}

func NewTemplate() (*Template, error) {
	tpl, err := template.New("tpl").Parse(`<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet">
			<title>HTTP Print Server</title>
		</head>
		<body>
			<div class="container">
				<nav class="navbar navbar-default">
					<div class="navbar-header">
						<a class="navbar-brand" href="/">HTTP Print Server</a>
					</div>
				</nav>
			<div class="panel panel-default">
				<div class="panel-heading">Printer {{.Printer}} - Jobs</div>
					<div class="panel-body">
						<table class="table">
							<tr>
								<th>Timestamp</th>
								<th>Content Type</th>
								<th>Length</th>
							</tr>
							{{range .Jobs }}
							<tr>
								<td>{{ .Timestamp }}</td>
								<td>{{ .ContentType }}</td>
								<td>{{ .Len }}</td>
							</tr>
							{{end}}
						</table>
					</div>
				</div>
			</div>
		</body>
		<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>
		<script type="text/javascript">
			setTimeout(function(){
			   window.location.reload(1);
				}, 2000);
		</script>
	</html>`)
	if err != nil {
		return nil, err
	}
	return &Template{
		tpl: tpl,
	}, err
}

func (t *Template) Render(w io.Writer, html_name string, data interface{}, c echo.Context) error {
	return t.tpl.Execute(w, data)
}

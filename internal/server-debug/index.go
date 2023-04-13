package serverdebug

import (
	"html/template"
	"strings"

	"github.com/karasunokami/chat-service/internal/logger"
	"github.com/labstack/echo/v4"
)

type page struct {
	Path        string
	Description string
}

type indexPage struct {
	pages []page
}

func newIndexPage() *indexPage {
	return &indexPage{}
}

func (i *indexPage) addPage(path string, description string) {
	i.pages = append(i.pages, page{
		Path:        path,
		Description: description,
	})
}

func (i indexPage) handler(eCtx echo.Context) error {
	return template.Must(template.New("index").Parse(`<html>
	<title>Chat Service Debug</title>
<body>
	<h2>Chat Service Debug</h2>
	<ul>
		{{range  .Pages}}
			 <li><a href="{{.Path}}">{{.Path}}</a> {{.Description}}</li>
		{{end}}
	</ul>

	<h2>Log Level</h2>
	<form onSubmit="putLogLevel()">
		<select id="log-level-select">
			{{ $l := .LogLevel }}
			{{range .LogLevels}}
				<option value="{{.}}" {{if eq $l .}}selected{{end}}>{{.}}</option>
			{{end}}
		</select>
		<input type="submit" value="Change"></input>
	</form>
	
	<script>
		function putLogLevel() {
			const req = new XMLHttpRequest();
			req.open('PUT', '/log/level', false);

			req.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');

			req.onload = function() { window.location.reload(); };
			req.send('level='+document.getElementById('log-level-select').value);
		};
	</script>
</body>
</html>
`)).Execute(eCtx.Response(), struct {
		Pages     []page
		LogLevel  string
		LogLevels []string
	}{
		Pages:     i.pages,
		LogLevels: []string{"DEBUG", "INFO", "WARN", "ERROR"},
		LogLevel:  strings.ToUpper(logger.Al.String()),
	})
}

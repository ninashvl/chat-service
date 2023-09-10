package serverdebug

import (
	"html/template"

	"github.com/labstack/echo/v4"

	"github.com/ninashvl/chat-service/internal/logger"
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
	return template.Must(template.New("index").Parse(`
<html>
   <title>Chat Service Debug</title>
   <body>
      <h2>Chat Service Debug</h2>
      <ul>
         {{range .Pages}}
         <li> <a href="{{.Path}}">{{.Path}}</a> {{.Description}}</li>
         {{end}}
      </ul>
      <h2>Log Level</h2>
      <form onSubmit="putLogLevel()">
         <select id="log-level-select">
            <option value="DEBUG">DEBUG</option>
            <option value="INFO">INFO</option>
            <option value="WARN">WARN</option>
            <option value="ERROR">ERROR</option>
         </select>
         <input type="submit" value="Change"></input>
      </form>
      <script>
         window.onload = function() {
               	getLogLevel();
           	};
         function getLogLevel() {
               const req = new XMLHttpRequest();
               req.open('GET', '/log/level', true);
               req.onload = function() {
                   if (req.status >= 200 && req.status < 400) {
                       document.getElementById('log-level-select').value = req.responseText;
                   } else {
                       console.error('Error: could not retrieve log level.');
                   }
               };
               req.send();
           }
      </script>
      <script>
         function putLogLevel() {
         	const req = new XMLHttpRequest();
         	req.open('PUT', '/log/level', false);
         	req.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
         	req.onload = function() { window.location.reload(); };
         	req.send('level='+document.getElementById('log-level-select').value);
         };
      </script>
   </body>
</html>
`)).Execute(eCtx.Response(), struct {
		Pages    []page
		LogLevel string
	}{
		Pages:    i.pages,
		LogLevel: logger.LogLevel(),
	})
}

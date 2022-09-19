package main

import (
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Define the template registry struct
type TemplateRegistry struct {
	templates *template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Renderer = &TemplateRegistry{
			templates: template.Must(template.ParseGlob("web_data/view/*.html")),
		}
		e.Router.Static("/css", "web_data/css")
		e.Router.Static("/js", "web_data/js")
		e.Router.Static("/img", "web_data/img")
		//https://github.com/BulmaTemplates/bulma-templates/blob/master/templates/landing.html
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "",
			Handler: func(c echo.Context) error {
				return c.Render(http.StatusOK, "frontpage.html", nil)
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

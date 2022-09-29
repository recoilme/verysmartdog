package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/rest"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/wesleym/telegramwidget"
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
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "",
			Handler: func(c echo.Context) error {
				// https://github.com/BulmaTemplates/bulma-templates/blob/master/templates/landing.html
				return c.Render(http.StatusOK, "frontpage.html", nil)
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/auth_tg_signup",
			Handler: func(c echo.Context) error {
				params, paramsErr := url.ParseQuery(c.Request().URL.RawQuery)
				if paramsErr != nil {
					return rest.NewBadRequestError("Failed to create user token, bad params", paramsErr)
				}
				uData, tgwErr := telegramwidget.ConvertAndVerifyForm(params, string("5537821699:AAFTg_0meVPkMrD-qY8kLSPkH6cXVaXcj1w"))
				if tgwErr != nil {
					return rest.NewBadRequestError("Failed to verify user token", tgwErr)
				}
				//uid := fmt.Sprintf("%d", u.ID)
				email := fmt.Sprintf("%d@t.me", uData.ID)
				user, userErr := app.Dao().FindUserByEmail(email)
				if userErr != nil {
					// not found user
					app.Dao().RunInTransaction(func(txDao *daos.Dao) error {

						user = &models.User{}
						user.Verified = false
						user.Email = email
						user.SetPassword(security.RandomString(30))

						// create the new user
						if err := txDao.SaveUser(user); err != nil {
							return err
						}

						return nil
					})
				}
				_ = user
				return c.Render(http.StatusOK, "frontpage.html", nil)
			},
		})

		return nil
	})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
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

func customAuthMiddleware(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Print(fmt.Sprintf("%+v\n", c.Request().URL.String()))
			tokenC, err := c.Cookie("t")
			if err != nil || tokenC == nil {
				if c.Request().URL.String() == "/" {
					//next(c)
					return c.Redirect(307, "frontpage")
				}
			} else {
				// set the user token to header for not admin urls
				//if !strings.HasPrefix(c.Request().URL.String(), "/_/") || !strings.HasPrefix(c.Request().URL.String(), "api/admins") {
				c.Request().Header.Set("Authorization", "User "+tokenC.Value)
				//}
			}
			return next(c)
		}
	}
}

func main() {
	app := pocketbase.New()

	//app.OnUserAfterCreateRequest().Add(func(e *core.UserCreateEvent) error {
	//	log.Println(e.User.Email)
	//	return nil
	//})

	//app.OnUserAuthRequest().Add(func(e *core.UserAuthEvent) error {
	//	log.Println(e.Token)
	//	return nil
	//})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Pre(customAuthMiddleware(app))
		e.Router.Renderer = &TemplateRegistry{
			templates: template.Must(template.ParseGlob("web_data/view/*.html")),
		}
		e.Router.Static("/css", "web_data/css")
		e.Router.Static("/js", "web_data/js")
		e.Router.Static("/img", "web_data/img")
		e.Router.HTTPErrorHandler = customHTTPErrorHandler
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {
				return c.Render(http.StatusOK, "main.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/frontpage",
			Handler: func(c echo.Context) error {
				return c.Render(http.StatusOK, "frontpage.html", nil)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/newfeed",
			Handler: func(c echo.Context) error {
				return c.Render(http.StatusOK, "newfeed.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/newfeed",
			Handler: func(c echo.Context) error {
				log.Println(fmt.Sprintf("post:%+v\n", c))
				link := c.FormValue("link")
				log.Println("link", link)
				return c.JSON(http.StatusOK, link)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/logout",
			Handler: func(c echo.Context) error {

				cookie := new(http.Cookie)
				cookie.Name = "t"
				cookie.Value = ""
				cookie.Expires = time.Now().Add((-1) * time.Second)
				c.SetCookie(cookie)

				return c.Redirect(307, "/")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/feed/:id/:name",
			Handler: func(c echo.Context) error {
				//log.Print(c.PathParams().Get("id", "-"), c.PathParams().Get("name", "-"))
				return c.Render(http.StatusOK, "main.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrUserAuth(),
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

func customHTTPErrorHandler(c echo.Context, err error) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	m := map[string]string{}
	m["code"] = strconv.Itoa(code)
	m["msg"] = err.Error()
	log.Print(err)
	c.Render(code, "404.html", m)

}

func siteData(c echo.Context) (siteData map[string]string) {
	siteData = map[string]string{}
	user, _ := c.Get(apis.ContextUserKey).(*models.User)
	siteData["photo_url"] = user.Profile.Data()["photo_url"].(string)
	siteData["name"] = user.Profile.Data()["name"].(string)
	siteData["userId"] = user.Id
	feedname := c.PathParams().Get("name", "-")
	if feedname == "-" {
		siteData["path"] = c.Request().URL.String()
		siteData["feedname"] = ""
	} else {
		siteData["path"] = "/feed"
		siteData["feedname"] = feedname
	}

	log.Println(fmt.Sprintf("siteData:%+v", siteData))
	return siteData
}

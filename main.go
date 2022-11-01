package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/recoilme/verysmartdog/internal/usecase"
	"github.com/recoilme/verysmartdog/internal/vsd"
	"github.com/spf13/cobra"
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
			log.Print(fmt.Sprintf("customAuthMiddleware: %+v\n", c.Request().URL.String()))
			tokenC, err := c.Cookie("t")
			if err != nil || tokenC == nil {
				if c.Request().URL.String() == "/" {
					return c.Redirect(http.StatusTemporaryRedirect, "/frontpage")
				}
			} else {
				c.Request().Header.Set("Authorization", "Bearer "+tokenC.Value)
			}
			return next(c)
		}
	}
}

func main() {
	app := pocketbase.New()

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
				apis.RequireAdminOrRecordAuth(),
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
				//feeds(c)
				log.Println("OneTimeErr2", c.Get("OneTimeErr"))
				return c.Render(http.StatusOK, "newfeed.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
			},
		})
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/newfeed",
			Handler: func(c echo.Context) error {
				link := c.FormValue("link")
				authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
				feeds, err := vsd.FeedNew(app, link, authRecord.GetId())
				_ = feeds
				if err != nil {
					c.Set("err", err.Error())
				}
				return c.Render(http.StatusOK, "newfeed.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
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

				return c.Redirect(http.StatusTemporaryRedirect, "/")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
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
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/auth_tg_signup",
			Handler: func(c echo.Context) error {
				authRecord, err := vsd.AuthTgSignup(app.Dao(), c.Request().URL.RawQuery)
				if err != nil {
					return apis.NewBadRequestError("Failed to auth", err)
				}

				token, tokenErr := tokens.NewRecordAuthToken(app, authRecord)
				if tokenErr != nil {
					return apis.NewBadRequestError("Failed to create auth token.", tokenErr)
				}
				cookie := new(http.Cookie)
				cookie.Name = "t"
				cookie.Value = token
				cookie.Expires = time.Now().Add(400 * 24 * time.Hour)
				c.SetCookie(cookie)

				return c.Redirect(307, "/")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})

		return nil
	})

	app.RootCmd.AddCommand(&cobra.Command{
		Use: "checkfeeds",
		Run: func(command *cobra.Command, args []string) {
			log.Println("Checking feeds started")
			updater := usecase.NewFeedsUpdater(app.DB())
			if err := updater.CheckFeeds(1); err != nil {
				log.Fatal(err)
			}
			log.Println("Checking feeds finished")
		},
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

func siteData(c echo.Context) (siteData map[string]interface{}) {
	siteData = map[string]interface{}{}
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

	siteData["photo_url"] = authRecord.GetString("photo_url")
	siteData["name"] = authRecord.GetString("username")
	siteData["userId"] = authRecord.GetId()
	feedname := c.PathParams().Get("name", "-")
	if feedname == "-" {
		siteData["path"] = c.Request().URL.String()
		siteData["feedname"] = ""
	} else {
		siteData["path"] = "/feed"
		siteData["feedname"] = feedname
	}
	siteData["feeds"] = c.Get("feeds")
	siteData["err"] = c.Get("err")

	log.Println(fmt.Sprintf("siteData:%+v", siteData))

	return siteData
}

func feeds(c echo.Context) {

	resp, err := http.Get("http://127.0.0.1:8090/api/collections/feed/records?page=1&perPage=10")
	if err != nil {
		log.Println("Err:", err)
		return
	}
	result := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Err:", err)
		return
	}
	c.Set("feeds", result["items"])
}

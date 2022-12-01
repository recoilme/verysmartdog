package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tokens"
	"github.com/recoilme/verysmartdog/internal/pagination"
	"github.com/recoilme/verysmartdog/internal/vsd"
	_ "github.com/recoilme/verysmartdog/migrations"
	"github.com/recoilme/verysmartdog/pkg/pbapi"
	"github.com/tidwall/interval"
)

// Define the template registry struct
type TemplateRegistry struct {
	templates *template.Template
}

var funcMap = template.FuncMap{
	"reltime": func(x string) string {
		t, err := time.Parse("2006-01-02 15:04:05.000Z", x)
		if err != nil {
			return x + " err:" + err.Error()
		}
		return humanize.RelTime(t, time.Now().UTC(), "ago", "later")
	},
}

func main() {
	botkey, err := os.ReadFile("tgbot")
	if err != nil {
		log.Fatal(err)
	}
	botkeys := string(bytes.TrimSpace(botkey))

	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.Pre(customAuthMiddleware(app))

		filenames, _ := filepath.Glob("web_data/view/*.html")
		e.Router.Renderer = &TemplateRegistry{
			templates: template.Must(template.New("main.html").Funcs(funcMap).ParseFiles(filenames...)),
		}
		e.Router.Static("/css", "web_data/css")
		e.Router.Static("/js", "web_data/js")
		e.Router.Static("/img", "web_data/img")
		e.Router.HTTPErrorHandler = customHTTPErrorHandler
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {
				usrFeeds(c, app)
				userId := ""
				if authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record); ok {
					userId = authRecord.GetId()
				}

				result, err := vsd.AllPosts(app, userId, c.QueryParam("page"))
				if err != nil {
					c.Set("err", err.Error())
				}
				c.Set("pagination", pagination.New(result.TotalItems, result.PerPage, result.Page))
				c.Set("posts", toJson(result)["items"])
				//c.Set("err", "In development")
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
				userId := ""
				if authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record); ok {
					userId = authRecord.GetId()
				}
				result, err := vsd.NotUserFeeds(app, userId)
				if err != nil {
					c.Set("err", err.Error())
				}

				bin, err := json.Marshal(result)
				if err != nil {
					fmt.Println(err)

				}
				resultJson := map[string]interface{}{}
				json.NewDecoder(bytes.NewReader(bin)).Decode(&resultJson)
				//log.Println("feeds", fmt.Sprintf("%+v\n", resultJson["items"]))
				c.Set("feeds", resultJson["items"])
				usrFeeds(c, app)
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
				usrFeeds(c, app)
				return c.Render(http.StatusOK, "newfeed.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				//apis.RequireAdminOrRecordAuth(),
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
			Path:   "/feed/:feedid/:domainname/:period",
			Handler: func(c echo.Context) error {
				//log.Print(c.PathParams().Get("id", "-"), c.PathParams().Get("name", "-"))
				usrFeeds(c, app)
				posts(c, app)
				return c.Render(http.StatusOK, "main.html", siteData(c))
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/feed/:feedid",
			Handler: func(c echo.Context) error {
				//log.Print(c.PathParams().Get("id", "-"), c.PathParams().Get("name", "-"))
				err := vsd.FeedUpd(app, c.PathParams().Get("feedid", "-"))
				if err != nil {
					log.Panicln("Err FeedUpd:", err.Error())
					return c.HTML(http.StatusInternalServerError, err.Error())
				}
				return c.HTML(http.StatusOK, "ok")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/subscriptions/:subscrtype/:subscrmethod/:subscrid",
			Handler: func(c echo.Context) error {
				//log.Print(c.PathParams().Get("subscrtype", "-"), c.PathParams().Get("subscrmethod", "-"))
				userId := ""
				authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
				if ok {
					userId = authRecord.GetId()
				}
				switch c.PathParams().Get("subscrtype", "-") {
				case "feed":
					switch c.PathParams().Get("subscrmethod", "-") {
					case "subscribe":
						err := vsd.SubscrFeed(app, c.PathParams().Get("subscrid", "-"), userId)
						if err != nil {
							return err
						}
					case "unsubscribe":
						err := vsd.UnsubscrFeed(app, c.PathParams().Get("subscrid", "-"), userId)
						if err != nil {
							return err
						}
					default:
						return errors.New("Unknown subscrmethod:" + c.PathParams().Get("subscrmethod", "-"))
					}
				default:
					return errors.New("Unknown subscrtype:" + c.PathParams().Get("subscrtype", "-"))
				}
				return c.Redirect(http.StatusTemporaryRedirect, "/")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireAdminOrRecordAuth(),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/auth_tg_signup",
			Handler: func(c echo.Context) error {
				authRecord, err := vsd.AuthTgSignup(app.Dao(), c.Request().URL.RawQuery, botkeys)
				if err != nil {
					log.Println("Failed to auth", err)
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

				return c.Redirect(http.StatusTemporaryRedirect, "/")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.RequireGuestOnly(),
			},
		})

		return nil
	})

	iv := interval.Set(func(t time.Time) {
		feedUpd(app)
	}, 1*time.Minute)

	if err := app.Start(); err != nil {
		log.Fatal(err)
		iv.Clear()
	}
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
				if strings.HasPrefix(c.Request().URL.String(), "/_/") ||
					strings.HasPrefix(c.Request().URL.String(), "/api/") {
					// do nothing with header for admins
				} else {
					c.Request().Header.Set("Authorization", "Bearer "+tokenC.Value)
				}
			}
			return next(c)
		}
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
	feedId := c.PathParams().Get("feedid", "-")
	if feedId == "-" {
		siteData["path"] = "/" //c.Request().URL.String()
		siteData["feedid"] = ""
	} else {
		siteData["path"] = "/feed"
		siteData["feedid"] = feedId
	}
	siteData["period"] = c.PathParams().Get("period", "-")
	siteData["domainname"] = c.PathParams().Get("domainname", "-")
	siteData["feeds"] = c.Get("feeds")
	siteData["err"] = c.Get("err")
	siteData["usr_feeds"] = c.Get("usr_feeds")
	siteData["posts"] = c.Get("posts")
	siteData["pagination"] = c.Get("pagination")
	//log.Println(fmt.Sprintf("siteData:%+v", siteData))

	return siteData
}

func usrFeeds(c echo.Context, app *pocketbase.PocketBase) {
	authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if ok {
		result, err := vsd.UsrFeeds(app, authRecord.GetId())
		if err != nil {
			c.Set("err", err.Error())
		}
		c.Set("usr_feeds", toJson(result)["items"])
	}
}

func posts(c echo.Context, app *pocketbase.PocketBase) {
	feedId := c.PathParams().Get("feedid", "-")
	period := c.PathParams().Get("period", "-")
	if feedId != "-" {
		result, err := vsd.Posts(app, feedId, period, c.QueryParam("page"))
		if err != nil {
			c.Set("err", err.Error())
		}
		c.Set("pagination", pagination.New(result.TotalItems, result.PerPage, result.Page))
		c.Set("posts", toJson(result)["items"])
	}
}

func toJson(in interface{}) map[string]interface{} {
	bin, err := json.Marshal(in)
	if err != nil {
		log.Println(err)
	}
	resultJson := map[string]interface{}{}
	json.NewDecoder(bytes.NewReader(bin)).Decode(&resultJson)
	return resultJson
}

func feedUpd(app *pocketbase.PocketBase) {
	hourAgo := time.Now().UTC().Add(-1 * time.Hour)

	sres, err := pbapi.RecordList(app, "feed", fmt.Sprintf("page=1&perPage=2&sort=last_fetch&filter=(last_fetch<'%s')", hourAgo.Format("2006-01-02 15:04:05")), "")
	if err != nil {
		log.Println(err)
		return
	}
	if x, ok := sres.Items.([]*models.Record); ok {
		for _, rec := range x {
			log.Println("fetch", rec.GetString("last_fetch"))
			err := vsd.FeedUpd(app, rec.Id)
			if err != nil {
				log.Println("Error: updating feed", rec.GetString("url"), err)
				return
			}
		}
	} else {
		fmt.Printf("I don't know how to handle %T\n", sres.Items)
	}
	return
}

package vsd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/SlyMarbo/rss"
	"github.com/katera/og"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/search"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/wesleym/telegramwidget"
)

func AuthTgSignup(dao *daos.Dao, queryParams string) (*models.Record, error) {

	params, paramsErr := url.ParseQuery(queryParams)
	if paramsErr != nil {
		return nil, apis.NewBadRequestError("Failed to create user token, bad params", paramsErr)
	}
	uData, tgwErr := telegramwidget.ConvertAndVerifyForm(params, string("5537821699:AAFTg_0meVPkMrD-qY8kLSPkH6cXVaXcj1w"))
	if tgwErr != nil {
		return nil, apis.NewBadRequestError("Failed to verify user token", tgwErr)
	}
	email := fmt.Sprintf("%d@t.me", uData.ID)
	authRecord, authRecordErr := dao.FindAuthRecordByEmail("users", email)
	if authRecordErr != nil {
		// not found user
		saveErr := dao.RunInTransaction(func(txDao *daos.Dao) error {

			collection, err := dao.FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}
			authRecord = models.NewRecord(collection)
			authRecord.SetEmail(email)
			authRecord.SetPassword(security.RandomString(30))
			authRecord.SetUsername(uData.Username)
			authRecord.Set("photo_url", uData.PhotoURL.String())

			// create the new user
			if err := txDao.Save(authRecord); err != nil {
				return err
			}

			return nil
		})
		if saveErr != nil {
			return nil, saveErr
		}
	}
	return authRecord, nil
}

func FeedNew(app core.App, link, userId string) ([]*models.Record, error) {

	linkUrl := link
	if !strings.HasPrefix(link, "http") {
		linkUrl = "http://" + link
	}
	hostname := ""
	fullUrl := ""
	uri, err := url.Parse(linkUrl)
	if err == nil {
		hostname = strings.TrimPrefix(uri.Hostname(), "www.")
		parts := strings.Split(hostname, ".")
		if len(parts) < 2 {
			hostname = ""
		}
		fullUrl = uri.Scheme + "://" + hostname
	}
	if hostname == "" {
		//search by link
		log.Println("search by link", err)
		return nil, errors.New("Err: no hostname in feed url:'" + link + "'")
	}
	feed, err := app.Dao().FindFirstRecordByData("feed", "url", link)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		// feed not found by url, try fetch
		fetchedFeed, err := rss.Fetch(link)
		if err != nil {
			return nil, err
		}

		domInfo, err := og.GetOpenGraphFromUrl("http://" + hostname)
		if err != nil {
			return nil, err
		}
		requestData := map[string]any{}
		if IsUrlValid(fullUrl) {
			requestData["url"] = fullUrl
		} else {
			if IsUrlValid(strings.TrimSuffix(domInfo.Url, "/")) {
				requestData["url"] = strings.TrimSuffix(domInfo.Url, "/")
			}
		}
		if IsUrlValid(requestData["url"].(string) + "/favicon.ico") {
			requestData["icon"] = requestData["url"].(string) + "/favicon.ico"
		} else {
			if len(domInfo.Images) > 0 {
				for _, img := range domInfo.Images {
					if IsUrlValid(img.URL) {
						requestData["icon"] = img.URL
						break
					}
				}
			}
		}
		requestData["hostname"] = hostname
		requestData["title"] = domInfo.Title
		requestData["descr"] = domInfo.Description
		requestData["lang"] = domInfo.Locale
		domain, err := CreateRecord(app, "domain", requestData)
		// feed
		requestData = map[string]any{}
		requestData["domain_id"] = domain.GetId()
		requestData["url"] = link
		requestData["title"] = fetchedFeed.Title
		requestData["descr"] = fetchedFeed.Description
		feed, err := CreateRecord(app, "feed", requestData)

		// usr_feed
		requestData = map[string]any{}
		requestData["user_id"] = userId
		requestData["feed_id"] = feed.GetId()
		_, err = CreateRecord(app, "usr_feed", requestData)
		return nil, err
	}
	// usr_feed
	requestData := map[string]any{}
	requestData["user_id"] = userId
	requestData["feed_id"] = feed.GetId()
	_, err = CreateRecord(app, "usr_feed", requestData)
	return nil, err
}

func IsUrlValid(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		return false
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

func Feeds(app core.App) (*search.Result, error) {
	return RecordsList(app, "feed", "", "domain_id")
}

func UsrFeeds(app core.App, userId string) (*search.Result, error) {
	return RecordsList(app, "usr_feed", fmt.Sprintf("filter=(user_id='%s')", userId), "feed_id,feed_id.domain_id")
}

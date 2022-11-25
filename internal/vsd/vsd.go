package vsd

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/katera/og"
	"github.com/mmcdole/gofeed"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/search"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/recoilme/verysmartdog/pkg/pbapi"
	"github.com/recoilme/verysmartdog/pkg/urls"
	"github.com/wesleym/telegramwidget"
)

func FeedNew(app core.App, link, userId string) ([]*models.Record, error) {
	link = strings.TrimSpace(link)
	domainUrl, hostname, err := urls.DomainHostName(link)
	if err != nil {
		log.Println("DomainHostName:", err)
		return nil, errors.New("Err: no hostname in feed url:'" + link + "'")
	}

	// domain
	requestData := map[string]any{}
	domain, err := app.Dao().FindFirstRecordByData("domain", "url", domainUrl)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// add domain
			requestData = map[string]any{}
			requestData["url"] = domainUrl
			requestData["hostname"] = hostname
			domInfo, err := og.GetOpenGraphFromUrl(domainUrl)
			if err != nil {
				return nil, err
			}
			requestData["title"] = domInfo.Title
			requestData["descr"] = domInfo.Description
			requestData["lang"] = domInfo.Locale
			favicon := strings.TrimRight(domainUrl, "/") + "/favicon.ico"
			if !urls.IsUrlValid(favicon) {
				if len(domInfo.Images) > 0 {
					for _, img := range domInfo.Images {
						if urls.IsUrlValid(img.URL) {
							favicon = img.URL
							break
						}
					}
				}
			}
			requestData["icon"] = favicon
			//requestData["lang"] = domInfo.Locale
			domain, err = pbapi.RecordCreate(app, "domain", &models.Admin{}, requestData)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	// feed
	feed, err := app.Dao().FindFirstRecordByData("feed", "url", link)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			// new feed
			//log.Println("new feed")
			fp := gofeed.NewParser()
			fetchedFeed, err := fp.ParseURL(link)
			if err != nil {
				log.Println("fetchedFeed", err)
				return nil, err
			}
			if fetchedFeed.Image != nil {
				domain.Set("icon", fetchedFeed.Image.URL)
				app.Dao().SaveRecord(domain)
			}
			requestData = map[string]any{}
			requestData["domain_id"] = domain.GetId()
			requestData["url"] = link
			requestData["title"] = fetchedFeed.Title
			requestData["descr"] = fetchedFeed.Description
			requestData["pub_date"] = fetchedFeed.UpdatedParsed
			requestData["lang"] = fetchedFeed.Language
			if fetchedFeed.Image != nil {
				requestData["icon"] = fetchedFeed.Image.URL
			}

			feed, err = pbapi.RecordCreate(app, "feed", &models.Admin{}, requestData)
			if err != nil {
				return nil, err
			}
		default:
			// some other error
			return nil, err
		}
	}
	_ = feed
	// usr_feed
	requestData = map[string]any{}
	requestData["user_id"] = userId
	requestData["feed_id"] = feed.GetId()
	_, err = pbapi.RecordCreate(app, "usr_feed", nil, requestData)
	if err != nil {
		return nil, err
	}
	go FeedUpd(app, feed.GetId())
	return nil, nil
}

func Feeds(app core.App) (*search.Result, error) {
	return pbapi.RecordList(app, "feed", "", "domain_id")
}

func UsrFeeds(app core.App, userId string) (*search.Result, error) {
	return pbapi.RecordList(app, "usr_feed", fmt.Sprintf("filter=(user_id='%s')", userId), "feed_id,feed_id.domain_id")
}

func Posts(app core.App, feedId, period string) (*search.Result, error) {
	td := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	tm := td.Add(24 * time.Hour)
	switch period {
	case "yesterday":
		td = td.Add(-24 * time.Hour)
		tm = tm.Add(-24 * time.Hour)
	case "week":
		td = td.Add(-24 * 7 * time.Hour)
	}
	query := fmt.Sprintf(`pub_date>="%s" && pub_date<"%s" && feed_id="%s"`,
		td.Format("2006-01-02"), tm.Format("2006-01-02"), feedId)
	//log.Println(query)

	filter := "sort=-pub_date&filter=" + url.QueryEscape(query)
	return pbapi.RecordList(app, "post", filter, "feed_id")
}

func FeedUpd(app core.App, feedId string) error {
	log.Println("FeedUpd", feedId)
	feed, err := app.Dao().FindFirstRecordByData("feed", "id", feedId)
	if err != nil {
		feed.Set("last_error", err.Error())
		feed.Set("last_fetch", time.Now())
		app.Dao().SaveRecord(feed)
		return err
	}
	fp := gofeed.NewParser()
	fetchedFeed, err := fp.ParseURL(feed.GetString("url"))
	if err != nil {
		feed.Set("last_error", err.Error())
		feed.Set("last_fetch", time.Now())
		app.Dao().SaveRecord(feed)
		return err
	}
	feed.Set("last_fetch", time.Now())
	feed.Set("last_error", "")
	feed.Set("pub_date", fetchedFeed.UpdatedParsed)
	app.Dao().SaveRecord(feed)
	for _, rssItem := range fetchedFeed.Items {
		title, err := goquery.NewDocumentFromReader(strings.NewReader(rssItem.Title))
		if err != nil {
			return err
		}
		titleTxt := title.Text()
		requestData := map[string]any{}
		requestData["feed_id"] = feed.GetId()
		requestData["url"] = rssItem.Link
		requestData["title"] = titleTxt
		requestData["pub_date"] = rssItem.PublishedParsed
		domInfo, err := og.GetOpenGraphFromUrl(rssItem.Link)
		if err != nil {
			return err
		}
		if domInfo.Description != "" {
			descrOg, err := goquery.NewDocumentFromReader(strings.NewReader(domInfo.Description))
			if err != nil {
				return err
			}
			requestData["descr"] = descrOg.Text()
		} else {
			summaryRss, err := goquery.NewDocumentFromReader(strings.NewReader(rssItem.Description))
			if err != nil {
				return err
			}
			requestData["descr"] = summaryRss.Text()
		}
		content := rssItem.Content
		if content == "" {
			content = rssItem.Description
		}
		requestData["sum_html"] = content
		contentRss, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			return err
		}
		requestData["sum_txt"] = contentRss.Text()
		if rssItem.Image != nil {
			requestData["img"] = rssItem.Image.URL
		}

		//log.Println(requestData)
		_, err = pbapi.RecordCreate(app, "post", &models.Admin{}, requestData)
		if err != nil {
			log.Println("FeedUpd", "RecordCreate", err)
			return err
		}
	}
	return nil
}

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

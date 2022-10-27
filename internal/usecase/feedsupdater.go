package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/pocketbase/dbx"
	"github.com/recoilme/verysmartdog/internal/model"
)

const (
	timeLayout = "2006-01-02 15:04:05.00"
)

type FeedsUpdater struct {
	db *dbx.DB
}

func NewFeedsUpdater(db *dbx.DB) *FeedsUpdater {
	return &FeedsUpdater{db: db}
}

func (u *FeedsUpdater) CheckFeeds(numWorkers int) error {
	var domains []*model.Domain
	err := u.db.NewQuery("SELECT * FROM domain").All(&domains)
	if err != nil {
		log.Fatal(err)
	}

	for _, domain := range domains {
		log.Printf("Processing domain id=%s (%s)", domain.ID, domain.Name)
		var feeds []*model.Feed
		err := u.db.NewQuery(fmt.Sprintf("SELECT * FROM feed WHERE domain_id = '%s'", domain.ID)).All(&feeds)
		if err != nil {
			return err
		}

		// @TODO: parallelize it
		for _, f := range feeds {
			log.Printf("Processing feed id=%s (%s)", f.ID, f.Title)
			updated, inserted, err := checkFeed(u.db, f)
			if err != nil {
				f.LastError = err.Error()
			}

			log.Printf("updated=%d posts, inserted=%d posts", updated, inserted)
			// update feed record at DB after all
			f.LastFetch = time.Now().Format(timeLayout)
			res, err := u.db.Update("feed", f.ToDBParams(true), dbx.NewExp(fmt.Sprintf("feed.id = '%s'", f.ID))).Execute()
			if err != nil {
				return err
			}
			_, err = res.RowsAffected()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkFeed(db *dbx.DB, f *model.Feed) (int, int, error) {
	// get previously saved feed posts ids (to determine insert or update)
	existPosts := make(map[string]struct{})
	rows, err := db.NewQuery(fmt.Sprintf("SELECT id FROM post WHERE feed_id = '%s'", f.ID)).Rows()
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var postID string
		err = rows.Scan(&postID)
		if err != nil {
			return 0, 0, err
		}
		existPosts[postID] = struct{}{}
	}

	// fetch posts from feed rss
	posts, err := fetchFeed(f)
	if err != nil {
		return 0, 0, err
	}
	updated, inserted := 0, 0
	// save posts to DB
	if posts != nil && len(posts) > 0 {
		// @TODO: use batch db operations
		for _, post := range posts {
			if _, isExists := existPosts[post.ID]; isExists {
				res, err := db.Update("post", post.ToDBParams(true), dbx.NewExp(fmt.Sprintf("post.id = %s", post.ID))).Execute()
				if err != nil {
					log.Fatal(err)
				}
				upd, err := res.RowsAffected()
				if err != nil {
					log.Fatal(err)
				}
				updated += int(upd)
			} else {
				_, err := db.Insert("post", post.ToDBParams(false)).Execute()
				if err != nil {
					log.Fatal(err)
				}
				inserted++
			}
		}
	}
	return updated, inserted, nil
}

func fetchFeed(feed *model.Feed) ([]*model.Post, error) {
	ts := time.Now()

	fetchedFeed, err := rss.Fetch(feed.URL)
	if err != nil {
		return nil, err
	}

	posts := make([]*model.Post, 0, len(fetchedFeed.Items))
	for _, item := range fetchedFeed.Items {
		posts = append(posts, &model.Post{
			ID:      item.ID,
			FeedID:  feed.ID,
			URL:     item.Link,
			Title:   item.Title,
			Descr:   item.Summary,
			Img:     item.Image.Href,
			SumHTML: item.Content,
			SumTXT:  item.Content,
			PubDate: item.Date.Format(timeLayout),
			Updated: ts.Format(timeLayout),
		})
	}

	return posts, nil
}

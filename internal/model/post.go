package model

import (
	"github.com/pocketbase/dbx"
)

type Post struct {
	ID      string
	FeedID  string
	URL     string
	Title   string
	Descr   string
	Img     string
	SumHTML string
	SumTXT  string
	PubDate string
	Created string
	Updated string
}

func (p *Post) ToDBParams(forUpdate bool) dbx.Params {
	px := dbx.Params{
		"id":       p.ID,
		"feed_id":  p.FeedID,
		"url":      p.URL,
		"title":    p.Title,
		"descr":    p.Descr,
		"img":      p.Img,
		"sum_html": p.SumHTML,
		"sum_txt":  p.SumTXT,
		"pub_date": p.PubDate,
		"updated":  p.Updated,
	}
	if !forUpdate {
		px["created"] = p.Created
	}
	return px
}

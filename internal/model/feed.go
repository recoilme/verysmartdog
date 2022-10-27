package model

import "github.com/pocketbase/dbx"

type Feed struct {
	ID       string
	DomainID string
	URL      string

	Title string
	Descr string

	LastFetch string
	LastError string
	RespCode  int
	Created   string
	Updated   string
}

func (p *Feed) ToDBParams(forUpdate bool) dbx.Params {
	px := dbx.Params{
		"id":        p.ID,
		"domain_id": p.DomainID,
		"url":       p.URL,
		"title":     p.Title,
		"descr":     p.Descr,

		"last_fetch": p.LastFetch,
		"last_error": p.LastError,
		"resp_code":  p.RespCode,
		"updated":    p.Updated,
	}
	if !forUpdate {
		px["created"] = p.Created
	}
	return px
}

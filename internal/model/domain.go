package model

import "github.com/pocketbase/dbx"

type Domain struct {
	ID   string
	URL  string
	Host string

	Name  string
	Title string
	Descr string

	Icon string
	Lang string

	Created string
}

func (p *Domain) ToDBParams(forUpdate bool) dbx.Params {
	px := dbx.Params{
		"id":    p.ID,
		"url":   p.URL,
		"host":  p.Host,
		"name":  p.Name,
		"title": p.Title,
		"descr": p.Descr,

		"icon": p.Icon,
		"lang": p.Lang,
	}
	if !forUpdate {
		px["created"] = p.Created
	}
	return px
}

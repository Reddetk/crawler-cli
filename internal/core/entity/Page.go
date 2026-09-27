// Package entity stands for busunes entities
package entity

import secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"

type Page struct {
	Resource string
	Title    string
	Links    []*Page
}

func NewPage() *Page {
	return &Page{}
}

func FormBlancPage(url string) *Page {
	return &Page{
		Resource: url,
	}
}

// Explore exploring Observe Results to page
func (pg *Page) Explore(obsRes secports.ObserveResults) []string {
	pg.Title = obsRes.Title
	return obsRes.Links
}

func (pg *Page) Adopt(childPg *Page) {
	pg.Links = append(pg.Links, childPg)
}

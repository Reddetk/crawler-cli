// Package entity stands for business entities
package entity

// Page is a node of the crawl result tree
type Page struct {
	Resource string  `json:"resource"`
	Title    string  `json:"title"`
	Links    []*Page `json:"links"`

	failed bool
}

// BlankPage creates an unexplored page by resource
func BlankPage(resource string) *Page {
	return &Page{Resource: resource, Links: []*Page{}}
}

// Adopt attaches the child page to the page
func (pg *Page) Adopt(child *Page) {
	pg.Links = append(pg.Links, child)
}

// Fail marks the page as not observed
func (pg *Page) Fail() {
	pg.failed = true
}

// Alive reports whether the page was observed successfully
func (pg *Page) Alive() bool {
	return !pg.failed
}

// Prune drops failed descendants from the subtree
func (pg *Page) Prune() {
	links := make([]*Page, 0, len(pg.Links))
	for _, child := range pg.Links {
		if !child.Alive() {
			continue
		}
		child.Prune()
		links = append(links, child)
	}
	pg.Links = links
}
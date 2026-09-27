package entity

import "net/url"

// CallTree is an aggregate of one start URL: root page, host border and visited urls
type CallTree struct {
	Root    *Page
	Host    string
	Visited *UniqMap
}

// NewCallTree builds a tree from the start URL
func NewCallTree(startURL string) (*CallTree, error) {
	u, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}
	return &CallTree{
		Root:    BlankPage(startURL),
		Host:    u.Host,
		Visited: NewUniqMap(),
	}, nil
}

// Allows reports whether the URL belongs to the tree host
func (ct *CallTree) Allows(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && u.Host == ct.Host
}

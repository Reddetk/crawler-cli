// Package entity contains domain entities for the crawler
package entity

import (
	"sync"
)

// URLTree tracks visited URLs within a single crawl tree (one root URL)
// to prevent duplicate requests and cyclic traversal.
type URLTree struct {
	mu       sync.Mutex
	observed map[string]bool
}

// NewURLTree creates an empty visited-URL tracker.
func NewURLTree() *URLTree {
	return &URLTree{
		observed: make(map[string]bool),
	}
}

// IsVisited reports whether the URL has already been observed.
// Use TryVisit instead when you need an atomic check-and-mark.
func (ut *URLTree) IsVisited(url string) bool {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	return ut.observed[url]
}

// MarkVisited records the URL as observed.
func (ut *URLTree) MarkVisited(url string) {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	ut.observed[url] = true
}

// TryVisit atomically checks whether the URL was already visited and,
// if not, marks it as visited.
func (ut *URLTree) TryVisit(url string) bool {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	if ut.observed[url] {
		return false
	}

	ut.observed[url] = true
	return true
}

// Count returns the number of unique URLs observed so far.
func (ut *URLTree) Count() int {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	return len(ut.observed)
}

// Snapshot returns a copy of all observed URLs.
func (ut *URLTree) Snapshot() []string {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	urls := make([]string, 0, len(ut.observed))
	for url := range ut.observed {
		urls = append(urls, url)
	}

	return urls
}

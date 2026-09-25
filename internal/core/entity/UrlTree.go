// Package entity contains domain entities for the crawler
package entity

import (
	"sync"
)

// URLTree tracks visited URLs
type URLTree struct {
	mu       sync.Mutex
	observed map[string]bool
}

func NewURLTree() *URLTree {
	return &URLTree{
		observed: make(map[string]bool),
	}
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

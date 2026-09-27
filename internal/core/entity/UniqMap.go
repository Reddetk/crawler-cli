package entity

import (
	"sync"
)

// UniqMap tracks visited URLs
type UniqMap struct {
	mu       sync.Mutex
	observed map[string]bool
}

func NewUniqMap() *UniqMap {
	return &UniqMap{
		observed: make(map[string]bool),
	}
}

// TryVisit atomically checks whether the URL was already visited and,
// if not, marks it as visited.
func (um *UniqMap) TryVisit(url string) bool {
	um.mu.Lock()
	defer um.mu.Unlock()

	if um.observed[url] {
		return false
	}

	um.observed[url] = true
	return true
}

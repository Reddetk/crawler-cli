package entity

import "sync"

// UniqMap tracks visited urls
type UniqMap struct {
	mu       sync.Mutex
	observed map[string]struct{}
}

// NewUniqMap creates an empty visited-urls map
func NewUniqMap() *UniqMap {
	return &UniqMap{observed: make(map[string]struct{})}
}

// TryVisit atomically reports whether the url is new and marks it visited
func (um *UniqMap) TryVisit(url string) bool {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, ok := um.observed[url]; ok {
		return false
	}
	um.observed[url] = struct{}{}
	return true
}

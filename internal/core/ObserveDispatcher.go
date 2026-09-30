package core

import (
	"context"

	"github.com/Reddetk/crawler-cli/cmd/config"
)

// ObserveDispatcher bounds the number of simultaneous observations
type ObserveDispatcher struct {
	sem chan struct{}
}

// NewObserveDispatcher creates a dispatcher with the given worker limit
func NewObserveDispatcher(maxWorkers int) (*ObserveDispatcher, error) {
	if err := config.ValidateMaxWorkers(maxWorkers); err != nil {
		return nil, err
	}
	return &ObserveDispatcher{sem: make(chan struct{}, maxWorkers)}, nil
}

// Acquire takes a free worker slot or fails with the context cancel
func (d *ObserveDispatcher) Acquire(ctx context.Context) error {
	select {
	case d.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release returns the worker slot
func (d *ObserveDispatcher) Release() {
	<-d.sem
}

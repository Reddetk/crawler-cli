package core

import "context"

type ObserveDispatcher struct {
	sem chan struct{}
}

func NewObserveDispatcher(maxWorkers int) *ObserveDispatcher {
	return &ObserveDispatcher{
		sem: make(chan struct{}, maxWorkers),
	}
}

func (d *ObserveDispatcher) Acquire(ctx context.Context) error {
	select {
	case d.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *ObserveDispatcher) Release() {
	<-d.sem
}

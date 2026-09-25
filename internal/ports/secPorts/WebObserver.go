// Package secports contain Secondary ports for service/Observer communicztion
package secports

import (
	"context"
)

type WebObserver interface {
	Observe(ctx context.Context, url string) (*ObserveResults, error)
}

type ObserveResults struct {
	Title string
	Links []string
}

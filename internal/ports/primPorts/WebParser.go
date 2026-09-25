// Package primports stands for web parse interfase
package primports

import "context"

type WebParser interface {
	StartCrawl(ctx context.Context, urls []string) error
}

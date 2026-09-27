// Package primports stands for web parse interfase
package primports

import (
	"context"

	"github.com/Reddetk/crawler-cli/internal/core/entity"
)

type WebParser interface {
	StartCrawl(ctx context.Context, urls []string) ([]*entity.Page, error)
}

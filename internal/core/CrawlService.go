// Package core conatain only busines ligic
package core

import (
	"context"
	"sync"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

const (
	chanBuffer int = 50
)

type CrawlerService struct {
	cnf         *config.ServiceConfig
	webObserver *secports.WebObserver
	log         logger.Logger
	urlTrees    []entity.URLTree
	mu          sync.Mutex
}

func (cs *CrawlerService) appendToUrlTree(urltree *entity.URLTree) {
}

func NewCrawlerService(
	cnf *config.ServiceConfig,
	webObserver *secports.WebObserver,
	log logger.Logger,
) *CrawlerService {
	reqChan := make(chan string, chanBuffer)
	return &CrawlerService{
		cnf:         cnf,
		webObserver: webObserver,
		log:         log,
		reqChan:     reqChan,
	}
}

func (cs *CrawlerService) StartCrawl(ctx context.Context, urls []string) error {
	URLTree := entity.NewURLTree()
}

func (cs *CrawlerService) processRequests(ctx context.Context, urls []string) error {
	return nil
}

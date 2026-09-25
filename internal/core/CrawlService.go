// Package core implement core busines logic CrawleService
package core

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

type CrawlerService struct {
	cnf         *config.ServiceConfig
	webObserver secports.WebObserver
	log         logger.Logger
	urlTree     *entity.URLTree
	dsp         *ObserveDispatcher
}

func NewCrawlerService(
	cnf *config.ServiceConfig,
	webObserver secports.WebObserver,
	log logger.Logger,
) *CrawlerService {
	return &CrawlerService{
		cnf:         cnf,
		webObserver: webObserver,
		log:         log,
		urlTree:     entity.NewURLTree(),
		dsp:         NewObserveDispatcher(cnf.AppEnvs.MaxWorkers),
	}
}

// StartCrawl start callObserverTree
func (cs *CrawlerService) StartCrawl(ctx context.Context, urls []string) error {
	ctx, cancelf := context.WithTimeout(ctx, cs.cnf.RequestTimeout)
	defer cancelf()

	g, ctx := errgroup.WithContext(ctx)

	for _, url := range urls {
		g.Go(func() error {
			return cs.crawl(ctx, g, url)
		})
	}

	return g.Wait()
}

// crawl one call Observer Leaf
func (cs *CrawlerService) crawl(ctx context.Context, g *errgroup.Group, depth int, url string) (*entity.Page, error) {
	if !cs.urlTree.TryVisit(url) {
		return nil, nil
	}

	if err := cs.dsp.Acquire(ctx); err != nil {
		return nil, err
	}
	defer cs.dsp.Release()

	ctxObs, cancelf := context.WithTimeout(ctx, cs.cnf.RequestTimeout)
	defer cancelf()

	res, err := cs.webObserver.Observe(ctxObs, url)
	if err != nil {
		cs.log.Warn("observe failed", logger.String("url", url), logger.Error(err))
		return nil, nil
	}

	if depth <= cs.cnf.Depth {
		return page, nil
	}

	for _, link := range res.Links {
		link := link
		g.Go(func() error {
			return cs.crawl(ctx, g, depth+1, link)
		})
	}

	return nil
}

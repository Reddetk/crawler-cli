// Package core implements the crawler business logic.
package core

import (
	"context"
	"sync"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

// CrawlerService orchestrates concurrent crawl trees over the web observer
type CrawlerService struct {
	cnf         *config.ServiceConfig
	webObserver secports.WebObserver
	dsp         *ObserveDispatcher
	log         logger.Logger
	trees       []*entity.CallTree
}

// NewCrawlerService creates a service with the bounded observer dispatcher
func NewCrawlerService(
	cnf *config.ServiceConfig,
	webObserver secports.WebObserver,
	log logger.Logger,
) *CrawlerService {
	return &CrawlerService{
		cnf:         cnf,
		webObserver: webObserver,
		dsp:         NewObserveDispatcher(cnf.AppEnvs.MaxWorkers),
		log:         log,
	}
}

// StartCrawl expands one call tree per start URL in parallel
func (cs *CrawlerService) StartCrawl(ctx context.Context, urls []string) ([]*entity.Page, error) {
	var wg sync.WaitGroup

	for _, u := range urls {
		tree, err := entity.NewCallTree(u)
		if err != nil {
			cs.log.Error("build call tree", logger.String("url", u), logger.Error(err))
			continue
		}
		if !tree.Visited.TryVisit(u) {
			continue
		}
		cs.trees = append(cs.trees, tree)

		wg.Add(1)
		go cs.appendTree(ctx, tree, tree.Root, 0, &wg)
	}

	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pages := make([]*entity.Page, 0, len(cs.trees))
	for _, t := range cs.trees {
		if !t.Root.Alive() {
			continue
		}
		t.Root.Prune()
		pages = append(pages, t.Root)
	}
	return pages, nil
}

// appendTree expands the page node inside the tree with fork-join
func (cs *CrawlerService) appendTree(ctx context.Context, tree *entity.CallTree, page *entity.Page, depth int, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := cs.dsp.Acquire(ctx); err != nil {
		return
	}
	defer cs.dsp.Release()

	links, err := cs.observe(ctx, page)
	if err != nil {
		return
	}
	if depth >= cs.cnf.Depth {
		return
	}

	for _, link := range links {
		if !tree.Allows(link) || !tree.Visited.TryVisit(link) {
			continue
		}
		child := entity.BlankPage(link)
		page.Adopt(child)

		wg.Add(1)
		go cs.appendTree(ctx, tree, child, depth+1, wg)
	}
}

// observe explores the page through the secondary port
func (cs *CrawlerService) observe(ctx context.Context, page *entity.Page) ([]string, error) {
	ctxReq, cancel := context.WithTimeout(ctx, cs.cnf.RequestTimeout)
	defer cancel()

	res, err := cs.webObserver.Observe(ctxReq, page.Resource)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		cs.log.Warn("observe failed", logger.String("url", page.Resource), logger.Error(err))
		page.Fail()
		return nil, nil
	}

	page.Title = res.Title
	return res.Links, nil
}

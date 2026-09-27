// Package core implement core business logic CrawlerService
package core

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

type CrawlerService struct {
	cnf         *config.ServiceConfig
	webObserver secports.WebObserver
	od          ObserveDispatcher
	log         logger.Logger
	callTrees   []*entity.CallTree
	mu          sync.Mutex
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
	}
}

type ObserveDispatcher struct {
	sym chan any
	wg  *sync.WaitGroup
}

func (cs *CrawlerService) newObserveDispatcher() *ObserveDispatcher {
	sym := make(chan any, cs.cnf.AppEnvs.MaxWorkers)
	wg := new(sync.WaitGroup)
	wg.Wait()
	return &ObserveDispatcher{
		sym: sym,
		wg:  wg,
	}
}

type Call struct {
	pg   *entity.Page
	dep  int
	host string
}

func NewRootCall(rturl string) (*Call, error) {
	rtBlancPage := entity.FormBlancPage(rturl)
	url, err := url.Parse(rturl)
	if err != nil {
		return nil, err
	}
	host := url.Host
	return &Call{
		pg:   rtBlancPage,
		dep:  0,
		host: host,
	}, nil
}

func (Call *Call) isHost() (bool, error) {
	url, err := url.Parse(Call.pg.Resource)
	if err != nil {
		return false, err
	}
	if url.Host != Call.pg.Resource {
		return false, nil
	}
	return true, nil
}

func (motherCall *Call) Child(childPg *entity.Page) *Call {
	return &Call{
		pg:   childPg,
		dep:  motherCall.dep + 1,
		host: motherCall.host,
	}
}

func (cs *CrawlerService) AppendToTree(ctx context.Context, caltree *entity.CallTree, curCall Call) (*Call, error) {
	obsRes, err := cs.webObserver.Observe(ctx, curCall.pg.Resource)
	if err != nil {
		return &curCall, fmt.Errorf("internal error: %w", err)
	}
	urls := curCall.pg.Explore(*obsRes)

	// stopers
	if curCall.dep >= cs.cnf.Depth {
		return &curCall, nil
	}

	for _, url := range urls {

		childPg := entity.FormBlancPage(url)
		curCall.pg.Adopt(childPg)
		if !caltree.Um.TryVisit(url) {
			return &curCall, nil
		}

		childCall := curCall.Child(childPg)
		ishst, err := childCall.isHost()
		if err != nil {
			return &curCall, fmt.Errorf("host definition error: %w", err)
		} else if !ishst {
			return &curCall, nil
		}
		cs.od.sym <- struct{}{}
		cs.od.wg.Add(1)

		go func() (*Call, error) {
			return cs.AppendToTree(ctx, caltree, *childCall)
		}()

		defer cs.od.wg.Done()
		<-cs.od.sym
	}
	return nil, nil
}

// StartCrawl обходит деревья всех стартовых URL параллельно.

func (cs *CrawlerService) StartCrawl(ctx context.Context, urls []string) ([]*entity.Page, error) {
	for _, url := range urls {
		rootCall, err := NewRootCall(url)
		if err != nil {
			cs.log.Error("new Call tree error: ", logger.Error(err))
			continue
		}
		caltree, err := entity.NewCallTree(rootCall.pg)
		if err != nil {
			cs.log.Error("new Call tree error: ", logger.Error(err))
			continue
		}
		cs.AddCallTree(caltree)
		_, err = cs.AppendToTree(ctx, caltree, *rootCall)
		if err != nil {
			cs.log.Error("new Call tree error: ", logger.Error(err))
			continue
		}
	}
}

func (cs *CrawlerService) AddCallTree(calltree *entity.CallTree) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.callTrees = append(cs.callTrees, calltree)
}

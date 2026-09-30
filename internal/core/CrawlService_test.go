package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

type nopLogger struct{}

func (nopLogger) Debug(string, ...logger.Field)      {}
func (nopLogger) Info(string, ...logger.Field)       {}
func (nopLogger) Warn(string, ...logger.Field)       {}
func (nopLogger) Error(string, ...logger.Field)      {}
func (nopLogger) With(...logger.Field) logger.Logger { return nopLogger{} }

type fakeObserver struct {
	results map[string]*secports.ObserveResults
	errs    map[string]error
	delays  map[string]time.Duration

	inFlight    atomic.Int64
	maxInFlight atomic.Int64
	onSuccess   func(url string)
}

var errBoom = errors.New("boom")

func (f *fakeObserver) Observe(ctx context.Context, url string) (*secports.ObserveResults, error) {
	cur := f.inFlight.Add(1)
	for {
		m := f.maxInFlight.Load()
		if cur <= m || f.maxInFlight.CompareAndSwap(m, cur) {
			break
		}
	}
	defer f.inFlight.Add(-1)

	if d, ok := f.delays[url]; ok && d > 0 {
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if err, ok := f.errs[url]; ok {
		return nil, err
	}
	if f.onSuccess != nil {
		f.onSuccess(url)
	}
	if r, ok := f.results[url]; ok {
		cp := *r
		return &cp, nil
	}
	return &secports.ObserveResults{}, nil
}

func newService(t *testing.T, obs secports.WebObserver, depth, maxWorkers int) *CrawlerService {
	t.Helper()

	t.Setenv("MAXWORKERS", strconv.Itoa(maxWorkers))
	cnf, err := config.InitDefaultServiceConfig()
	if err != nil {
		t.Error(err)
	}
	cnf.Depth = depth
	cnf.RequestTimeout = 5 * time.Second

	crwSrv, err := NewCrawlerService(cnf, obs, nopLogger{})
	if err != nil {
		t.Error(err)
	}

	return crwSrv
}

func countResources(pg *entity.Page, acc map[string]int) {
	acc[pg.Resource]++
	for _, c := range pg.Links {
		countResources(c, acc)
	}
}

func find(pg *entity.Page, resource string) *entity.Page {
	if pg.Resource == resource {
		return pg
	}
	for _, c := range pg.Links {
		if found := find(c, resource); found != nil {
			return found
		}
	}
	return nil
}

func TestStartCrawl_BuildsTreeUpToDepth(t *testing.T) {
	obs := &fakeObserver{results: map[string]*secports.ObserveResults{
		"https://site.com/a": {Links: []string{"https://site.com/b"}},
		"https://site.com/b": {Links: []string{"https://site.com/c"}},
		"https://site.com/c": {Links: []string{"https://site.com/d"}},
	}}
	svc := newService(t, obs, 2, 4)

	pages, err := svc.StartCrawl(context.Background(), []string{"https://site.com/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := find(pages[0], "https://site.com/c")
	if c == nil || len(c.Links) != 0 {
		t.Fatalf("depth cut failed: c must be a leaf")
	}
	if find(pages[0], "https://site.com/d") != nil {
		t.Fatalf("d must not be crawled at depth 2")
	}
}

func TestStartCrawl_SkipsDuplicatesAndCycles(t *testing.T) {
	obs := &fakeObserver{results: map[string]*secports.ObserveResults{
		"https://site.com/a": {Links: []string{"https://site.com/a", "https://site.com/b"}},
		"https://site.com/b": {Links: []string{"https://site.com/a", "https://site.com/b"}},
	}}
	svc := newService(t, obs, 5, 4)

	pages, err := svc.StartCrawl(context.Background(), []string{"https://site.com/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	acc := map[string]int{}
	countResources(pages[0], acc)
	if acc["https://site.com/a"] != 1 || acc["https://site.com/b"] != 1 {
		t.Fatalf("duplicate urls in tree: %v", acc)
	}
}

func TestStartCrawl_FiltersForeignHost(t *testing.T) {
	obs := &fakeObserver{results: map[string]*secports.ObserveResults{
		"https://site.com/a": {Links: []string{"https://site.com/b", "https://other.com/x"}},
	}}
	svc := newService(t, obs, 3, 4)

	pages, err := svc.StartCrawl(context.Background(), []string{"https://site.com/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	acc := map[string]int{}
	countResources(pages[0], acc)
	if _, ok := acc["https://other.com/x"]; ok {
		t.Fatalf("foreign host must not be crawled")
	}
}

func TestStartCrawl_ResourceErrorSkipsPage(t *testing.T) {
	obs := &fakeObserver{
		results: map[string]*secports.ObserveResults{
			"https://site.com/a": {Links: []string{"https://site.com/b", "https://site.com/c"}},
		},
		errs: map[string]error{"https://site.com/b": errors.New("boom")},
	}
	svc := newService(t, obs, 3, 4)

	pages, err := svc.StartCrawl(context.Background(), []string{"https://site.com/a"})
	if err != nil {
		t.Fatalf("resource error must not fail the crawl: %v", err)
	}
	if find(pages[0], "https://site.com/b") != nil {
		t.Fatalf("failed page must be pruned")
	}
	if find(pages[0], "https://site.com/c") == nil {
		t.Fatalf("sibling page must be crawled")
	}
}

func TestStartCrawl_CancelReturnsPartialResult(t *testing.T) {
	observed := make(chan string, 4)
	obs := &fakeObserver{
		results: map[string]*secports.ObserveResults{
			"https://site.com/a": {Links: []string{"https://site.com/slow"}},
		},
		delays:    map[string]time.Duration{"https://site.com/slow": 500 * time.Millisecond},
		onSuccess: func(url string) { observed <- url },
	}
	svc := newService(t, obs, 3, 4)

	ctx, cancel := context.WithCancel(context.Background())
	var (
		pages []*entity.Page
		err   error
		done  = make(chan struct{})
	)
	go func() {
		pages, err = svc.StartCrawl(ctx, []string{"https://site.com/a"})
		close(done)
	}()

	<-observed
	cancel()
	<-done

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if len(pages) != 1 || pages[0].Resource != "https://site.com/a" {
		t.Fatalf("partial result must contain the observed root")
	}
	if len(pages[0].Links) != 0 {
		t.Fatalf("unobserved child must be pruned")
	}
}

func TestStartCrawl_TotalTimeout(t *testing.T) {
	obs := &fakeObserver{
		delays: map[string]time.Duration{"https://site.com/a": 200 * time.Millisecond},
	}
	svc := newService(t, obs, 3, 4)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	pages, err := svc.StartCrawl(ctx, []string{"https://site.com/a"})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want context.DeadlineExceeded, got %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("nothing can be observed before the deadline")
	}
}

func TestStartCrawl_BoundsConcurrency(t *testing.T) {
	links := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		links = append(links, fmt.Sprintf("https://site.com/p%d", i))
	}

	results := map[string]*secports.ObserveResults{
		"https://site.com/a": {Links: links},
	}
	delays := map[string]time.Duration{}
	for _, l := range links {
		results[l] = &secports.ObserveResults{}
		delays[l] = 30 * time.Millisecond
	}

	obs := &fakeObserver{results: results, delays: delays}
	svc := newService(t, obs, 2, 2)

	pages, err := svc.StartCrawl(context.Background(), []string{"https://site.com/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages[0].Links) != 10 {
		t.Fatalf("incomplete crawl")
	}
	if got := obs.maxInFlight.Load(); got > 2 {
		t.Fatalf("concurrency bound violated: %d in flight", got)
	}
}

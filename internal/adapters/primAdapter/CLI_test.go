package primadapter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
)

type nopLogger struct{}

func (nopLogger) Debug(string, ...logger.Field)      {}
func (nopLogger) Info(string, ...logger.Field)       {}
func (nopLogger) Warn(string, ...logger.Field)       {}
func (nopLogger) Error(string, ...logger.Field)      {}
func (nopLogger) With(...logger.Field) logger.Logger { return nopLogger{} }

type fakeParser struct {
	pages []*entity.Page
	err   error
	crawl func(ctx context.Context, urls []string) ([]*entity.Page, error)
}

func (f *fakeParser) StartCrawl(ctx context.Context, urls []string) ([]*entity.Page, error) {
	if f.crawl != nil {
		return f.crawl(ctx, urls)
	}
	return f.pages, f.err
}

func newTestCLI(t *testing.T, p *fakeParser, timeout time.Duration) (*CLI, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "result.json")
	appCnf := config.InitDefaultAppConfig()
	appCnf.ResultPath = path
	appCnf.AppTimeout = timeout

	cli := NewCLI()
	cli.urls = []string{"https://site.com/a"}
	cli.Bind(p, appCnf, nopLogger{})
	return cli, path
}

func assertFileContains(t *testing.T, path, resource string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("result file not written: %v", err)
	}
	var pages []*entity.Page
	if err := json.Unmarshal(data, &pages); err != nil {
		t.Fatalf("invalid json %q: %v", data, err)
	}
	for _, p := range pages {
		if p.Resource == resource {
			return
		}
	}
	t.Fatalf("resource %s not found in %s", resource, data)
}

func TestRun_Success(t *testing.T) {
	p := &fakeParser{pages: []*entity.Page{entity.BlankPage("https://site.com/a")}}
	cli, path := newTestCLI(t, p, time.Minute)

	if outcome := cli.Run(context.Background()); outcome != OutcomeSuccess || outcome.ExitCode() != 0 {
		t.Fatalf("want success/0, got %v/%d", outcome, outcome.ExitCode())
	}
	assertFileContains(t, path, "https://site.com/a")
}

func TestRun_InterruptedKeepsPartialResult(t *testing.T) {
	p := &fakeParser{
		pages: []*entity.Page{entity.BlankPage("https://site.com/a")},
		err:   context.Canceled,
	}
	cli, path := newTestCLI(t, p, time.Minute)

	if outcome := cli.Run(context.Background()); outcome != OutcomeInterrupted || outcome.ExitCode() != 130 {
		t.Fatalf("want interrupted/130, got %v/%d", outcome, outcome.ExitCode())
	}
	assertFileContains(t, path, "https://site.com/a")
}

func TestRun_TimeoutKeepsPartialResult(t *testing.T) {
	p := &fakeParser{
		pages: []*entity.Page{entity.BlankPage("https://site.com/a")},
		err:   context.DeadlineExceeded,
	}
	cli, path := newTestCLI(t, p, time.Minute)

	if outcome := cli.Run(context.Background()); outcome != OutcomeTimeout || outcome.ExitCode() != 1 {
		t.Fatalf("want timeout/1, got %v/%d", outcome, outcome.ExitCode())
	}
	assertFileContains(t, path, "https://site.com/a")
}

func TestRun_AppliesAppTimeout(t *testing.T) {
	p := &fakeParser{crawl: func(ctx context.Context, urls []string) ([]*entity.Page, error) {
		<-ctx.Done() // как реальный сервис: работа до отмены бюджета
		return nil, ctx.Err()
	}}
	cli, _ := newTestCLI(t, p, 20*time.Millisecond)

	done := make(chan Outcome, 1)
	go func() { done <- cli.Run(context.Background()) }()

	select {
	case outcome := <-done:
		if outcome != OutcomeTimeout {
			t.Fatalf("want timeout, got %v", outcome)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("app timeout was not applied to the crawl context")
	}
}

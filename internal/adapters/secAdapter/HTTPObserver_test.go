package secadapter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/encoding/charmap"

	"github.com/Reddetk/crawler-cli/cmd/logger"
)

type nopLogger struct{}

func (nopLogger) Debug(string, ...logger.Field)      {}
func (nopLogger) Info(string, ...logger.Field)       {}
func (nopLogger) Warn(string, ...logger.Field)       {}
func (nopLogger) Error(string, ...logger.Field)      {}
func (nopLogger) With(...logger.Field) logger.Logger { return nopLogger{} }

func TestObserve_ExtractsTitleAndLinks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><head><title>Test page</title></head><body>
			<a href="https://site.com/b">b</a><a href="https.com/c">c</a></body></html>`)
	}))
	defer srv.Close()

	res, err := NewHTTPObserver(nopLogger{}).Observe(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Test page" {
		t.Fatalf("want title 'Test page', got %q", res.Title)
	}
	if len(res.Links) != 2 {
		t.Fatalf("want 2 links, got %v", res.Links)
	}
}

func TestObserve_SkipsRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://site.com/other", http.StatusMovedPermanently)
	}))
	defer srv.Close()

	if _, err := NewHTTPObserver(nopLogger{}).Observe(context.Background(), srv.URL); err == nil {
		t.Fatal("redirect must be skipped with an error")
	}
}

func TestObserve_SkipsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(http.NotFound))
	defer srv.Close()

	if _, err := NewHTTPObserver(nopLogger{}).Observe(context.Background(), srv.URL); err == nil {
		t.Fatal("404 must be skipped with an error")
	}
}

func TestObserve_SkipsNonHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		fmt.Fprint(w, "%PDF-1.4 fake")
	}))
	defer srv.Close()

	if _, err := NewHTTPObserver(nopLogger{}).Observe(context.Background(), srv.URL); err == nil {
		t.Fatal("non-html content must be skipped with an error")
	}
}

func TestObserve_ConvertsLegacyCharset(t *testing.T) {
	src := `<html><head><title>Тестовая страница</title></head><body><a href="https://site.com/b">b</a></body></html>`
	body, err := charmap.Windows1251.NewEncoder().Bytes([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=windows-1251")
		w.Write(body)
	}))
	defer srv.Close()

	res, err := NewHTTPObserver(nopLogger{}).Observe(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Тестовая страница" {
		t.Fatalf("cp1251 title must be decoded, got %q", res.Title)
	}
}

func TestObserve_RespectsCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><title>x</title></html>`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := NewHTTPObserver(nopLogger{}).Observe(ctx, srv.URL); err == nil {
		t.Fatal("cancelled context must fail the observe")
	}
}

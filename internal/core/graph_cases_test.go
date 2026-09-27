package core

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Reddetk/crawler-cli/internal/core/entity"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

func graphObserver(pages map[string][]string, errs map[string]error) *fakeObserver {
	results := make(map[string]*secports.ObserveResults, len(pages))
	for url, links := range pages {
		results[url] = &secports.ObserveResults{Links: links}
	}
	return &fakeObserver{results: results, errs: errs}
}

func crawlWithDeadline(t *testing.T, svc *CrawlerService, urls []string) ([]*entity.Page, error) {
	t.Helper()

	type outcome struct {
		pages []*entity.Page
		err   error
	}
	done := make(chan outcome, 1)
	go func() {
		p, err := svc.StartCrawl(context.Background(), urls)
		done <- outcome{p, err}
	}()

	select {
	case r := <-done:
		return r.pages, r.err
	case <-time.After(5 * time.Second):
		t.Fatal("crawl did not terminate: cycle was not broken")
	}
	return nil, nil
}

func TestStartCrawl_CyclicGraphs(t *testing.T) {
	cases := []struct {
		name    string
		pages   map[string][]string
		errs    map[string]error
		roots   []string
		depth   int
		workers int
		want    map[string]int
	}{
		{
			name:    "self loop",
			pages:   map[string][]string{"https://s.io/a": {"https://s.io/a"}},
			roots:   []string{"https://s.io/a"},
			depth:   3,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1},
		},
		{
			name:    "two node cycle",
			pages:   map[string][]string{"https://s.io/a": {"https://s.io/b"}, "https://s.io/b": {"https://s.io/a"}},
			roots:   []string{"https://s.io/a"},
			depth:   3,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1},
		},
		{
			name: "duplicate edges both directions",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/a", "https://s.io/b"},
				"https://s.io/b": {"https://s.io/a", "https://s.io/b"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   3,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1},
		},
		{
			name: "triangle",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b"},
				"https://s.io/b": {"https://s.io/c"},
				"https://s.io/c": {"https://s.io/a"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   5,
			workers: 3,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1, "https://s.io/c": 1},
		},
		{
			name: "diamond with shared target",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b", "https://s.io/c"},
				"https://s.io/b": {"https://s.io/d"},
				"https://s.io/c": {"https://s.io/d"},
				"https://s.io/d": {"https://s.io/a"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   5,
			workers: 3,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1, "https://s.io/c": 1, "https://s.io/d": 1},
		},
		{
			name: "four node ring",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b"},
				"https://s.io/b": {"https://s.io/c"},
				"https://s.io/c": {"https://s.io/d"},
				"https://s.io/d": {"https://s.io/a"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   5,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1, "https://s.io/c": 1, "https://s.io/d": 1},
		},
		{
			name: "repeated edge and cross links",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b", "https://s.io/b", "https://s.io/c"},
				"https://s.io/b": {"https://s.io/c"},
				"https://s.io/c": {"https://s.io/b"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   5,
			workers: 3,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1, "https://s.io/c": 1},
		},
		{
			name: "two roots mutual cycle: per-tree dedup",
			pages: map[string][]string{
				"https://s.io/x": {"https://s.io/y"},
				"https://s.io/y": {"https://s.io/x"},
			},
			roots:   []string{"https://s.io/x", "https://s.io/y"},
			depth:   3,
			workers: 2,
			want:    map[string]int{"https://s.io/x": 2, "https://s.io/y": 2},
		},
		{
			name: "depth cut closes the cycle early",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b"},
				"https://s.io/b": {"https://s.io/c"},
				"https://s.io/c": {"https://s.io/a"},
			},
			roots:   []string{"https://s.io/a"},
			depth:   1,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/b": 1},
		},
		{
			name: "cycle with failed node",
			pages: map[string][]string{
				"https://s.io/a": {"https://s.io/b", "https://s.io/c"},
				"https://s.io/c": {"https://s.io/a"},
			},
			errs:    map[string]error{"https://s.io/b": errBoom},
			roots:   []string{"https://s.io/a"},
			depth:   3,
			workers: 2,
			want:    map[string]int{"https://s.io/a": 1, "https://s.io/c": 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(t, graphObserver(tc.pages, tc.errs), tc.depth, tc.workers)

			pages, err := crawlWithDeadline(t, svc, tc.roots)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := map[string]int{}
			for _, p := range pages {
				countResources(p, got)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("tree mismatch:\n got:  %v\n want: %v", got, tc.want)
			}
		})
	}
}

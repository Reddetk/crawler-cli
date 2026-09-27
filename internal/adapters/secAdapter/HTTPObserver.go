package secadapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"github.com/Reddetk/crawler-cli/cmd/logger"
	secports "github.com/Reddetk/crawler-cli/internal/ports/secPorts"
)

type HTTPObserver struct {
	client *http.Client
	log    logger.Logger
}

func NewHTTPObserver(log logger.Logger) *HTTPObserver {
	return &HTTPObserver{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		log: log,
	}
}

func (o *HTTPObserver) Observe(ctx context.Context, url string) (*secports.ObserveResults, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/html")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		return nil, fmt.Errorf("non-html content-type: %q", ct)
	}

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("charset: %w", err)
	}

	return parseHTML(r)
}

func parseHTML(r io.Reader) (*secports.ObserveResults, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	res := &secports.ObserveResults{}
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && res.Title == "" {
					res.Title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "a":
				for _, a := range n.Attr {
					if a.Key == "href" {
						res.Links = append(res.Links, a.Val)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return res, nil
}

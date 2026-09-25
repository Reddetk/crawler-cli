// Package primadapter stay for CLI interaction model
package primadapter

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Reddetk/crawler-cli/cmd/logger"
	primports "github.com/Reddetk/crawler-cli/internal/ports/primPorts"
)

const (
	defaultRequepstTimeout time.Duration = time.Minute
	defaultDepth           int           = 3
)

type CLI struct {
	log                 logger.Logger
	WebParser           primports.WebParser
	ServiceConfigurator primports.ServiceConfigurator
}

func NewCLI(webParser primports.WebParser, serviceConfigurator primports.ServiceConfigurator, log logger.Logger) *CLI {
	return &CLI{
		log:                 log,
		WebParser:           webParser,
		ServiceConfigurator: serviceConfigurator,
	}
}

func (cli *CLI) ParseConsoleInput() {
}

func (cli *CLI) parseServiceConfigFlags() (int, time.Duration, error) {
	fs := flag.NewFlagSet("service-configs", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // подавить автоматический usage-вывод в stderr

	reqTimeout := flag.Duration("request-timeout", defaultRequepstTimeout, "Timeout for a single HTTP request, e.g. 10s, 500ms. Format: Go time.Duration")
	depth := flag.Int("depth", defaultDepth, "Maximum link-following depth from each start URL (0 = only the start URL itself)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		err := fmt.Errorf("failed to parse flags %w", err)
		return defaultDepth, defaultRequepstTimeout, err
	}

	return *depth, *reqTimeout, nil
}

func (cli *CLI) getPlayload() ([]string, error) {
	fs := flag.NewFlagSet("playload", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // подавить автоматический usage-вывод в stderr

	urls := flag.String("urls", "", "Comma-separated list of start URLs. Crawling for each URL is restricted to its own domain")

	if err := fs.Parse(os.Args[1:]); err != nil {
		err := fmt.Errorf("failed to parse flags %w", err)
		return nil, err
	}

	return parseUrls(*urls)
}

func (cli *CLI) processRequests(rootCtx context.Context, urls []string) {
	for _, url := range urls {
		go cli.WebParser.ProcessRequest(rootCtx, url)
	}
}

// helper
func parseUrls(url string) ([]string, error) {
	urls := strings.Split(url, ",")
	if len(urls) == 0 {
		return nil, fmt.Errorf("no url parsed")
	}
	return urls, nil
}

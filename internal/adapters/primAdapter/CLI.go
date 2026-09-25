// Package primadapter stay for SeedProduser interaction model
package primadapter

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Reddetk/crawler-cli/cmd/config"
	primports "github.com/Reddetk/crawler-cli/internal/ports/primPorts"
)

type SeedProduser struct {
	WebParser primports.WebParser
}

type CLI struct {
	urls     []string
	seedprod *SeedProduser
}

func NewSeedProduser(webparse primports.WebParser) *SeedProduser {
	return &SeedProduser{
		WebParser: webparse,
	}
}

func NewCLI() *CLI {
	return &CLI{}
}

func (cli *CLI) StartSeedProdusing(ctx context.Context) error {
	if cli.urls == nil {
		return fmt.Errorf("urls is empty")
	}
	if cli.seedprod == nil {
		return fmt.Errorf("seed producer not initialized")
	}

	err := cli.seedprod.WebParser.StartCrawl(ctx, cli.urls)
	if err != nil {
		return err
	}
	return nil
}

// ParseFlags get flags
// If val for config doesn't set, leave default
func (cli *CLI) ParseFlags(cnf *config.AppConfig, srvCnf *config.ServiceConfig) (*config.AppConfig, *config.ServiceConfig, error) {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // подавить автоматический usage-вывод в stderr

	var errs []error

	// App Configs
	timeout := fs.Duration("timeout", cnf.AppTimeout,
		"overall timeout for the whole crawl run, e.g. 2m, 90s, 1h30m. Format: Go time.Duration")
	output := fs.String("output", cnf.ResultPath, "result output path")
	logPath := fs.String("log", "", "additional logger output path")

	cnf.AppTimeout = *timeout

	if err := validatePath(*output); err != nil {
		errs = append(errs, fmt.Errorf("output: %w", err))
	} else {
		cnf.ResultPath = *output
	}

	if *logPath != "" {
		if err := validatePath(*logPath); err != nil {
			errs = append(errs, fmt.Errorf("log: %w", err))
		} else {
			cnf.LogCnf.WithOutputPath(*logPath)
		}
	}

	// Service configs

	reqTimeout := fs.Duration("request-timeout", srvCnf.RequestTimeout, "Timeout for a single HTTP request, e.g. 10s, 500ms. Format: Go time.Duration")
	depth := fs.Int("depth", srvCnf.Depth, "Maximum link-following depth from each start URL (0 = only the start URL itself)")

	srvCnf.RequestTimeout = *reqTimeout
	srvCnf.Depth = *depth

	// Playload

	urlsStr := fs.String("urls", "", "Comma-separated list of start URLs. Crawling for each URL is restricted to its own domain")
	urls, err := parseUrls(*urlsStr)

	cli.urls = urls

	if err := fs.Parse(os.Args[1:]); err != nil {
		err := fmt.Errorf("failed to parse flags %w", err)
		return cnf, srvCnf, err
	}

	errs = append(errs, err)
	return cnf, srvCnf, errors.Join(errs...)
}

// helper
func parseUrls(url string) ([]string, error) {
	urls := strings.Split(url, ",")
	if len(urls) == 0 {
		return nil, fmt.Errorf("no url parsed")
	}
	return urls, nil
}

// helpers
func validatePath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("cannot stat path %s: %w", path, err)
	}
	return nil
}

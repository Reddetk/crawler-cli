// Package primadapter stays for CLI interaction model.
package primadapter

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	primports "github.com/Reddetk/crawler-cli/internal/ports/primPorts"
)

type SeedProducer struct {
	WebParser primports.WebParser
	appCnf    *config.AppConfig
}

type CLI struct {
	urls     []string
	seedprod *SeedProducer
}

func NewCLI() *CLI {
	return &CLI{}
}

// Bind attaches the web parser and app config to the CLI.
func (cli *CLI) Bind(p primports.WebParser, appCnf *config.AppConfig) {
	cli.seedprod = &SeedProducer{WebParser: p, appCnf: appCnf}
}

// StartSeedProdusing runs the crawl bounded by the overall app timeout.
func (cli *CLI) StartSeedProdusing(ctx context.Context) ([]*entity.Page, error) {
	if len(cli.urls) == 0 {
		return nil, fmt.Errorf("urls is empty")
	}
	if cli.seedprod == nil {
		return nil, fmt.Errorf("seed producer not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, cli.seedprod.appCnf.AppTimeout)
	defer cancel()

	return cli.seedprod.WebParser.StartCrawl(ctx, cli.urls)
}

// ParseFlags fills configs and payload from console flags.
func (cli *CLI) ParseFlags(cnf *config.AppConfig, srvCnf *config.ServiceConfig) (*config.AppConfig, *config.ServiceConfig, error) {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var errs []error

	timeout := fs.Duration("timeout", cnf.AppTimeout, "overall timeout for the whole crawl run, e.g. 2m, 90s, 1h30m")
	output := fs.String("output", cnf.ResultPath, "result output path")
	logPath := fs.String("log", "", "additional logger output path")
	reqTimeout := fs.Duration("request-timeout", srvCnf.RequestTimeout, "timeout for a single HTTP request")
	depth := fs.Int("depth", srvCnf.Depth, "maximum link-following depth from each start URL (0 = only the start URL itself)")
	urlsStr := fs.String("urls", "", "comma-separated list of start URLs")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cnf, srvCnf, fmt.Errorf("failed to parse flags %w", err)
	}

	cnf.AppTimeout = *timeout
	srvCnf.RequestTimeout = *reqTimeout
	srvCnf.Depth = *depth

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

	urls, err := parseUrls(*urlsStr)
	if err != nil {
		errs = append(errs, err)
	} else {
		cli.urls = urls
	}

	return cnf, srvCnf, errors.Join(errs...)
}

func parseUrls(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("no urls provided")
	}
	return strings.Split(raw, ","), nil
}

func validatePath(path string) error {
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("output directory does not exist: %s", dir)
	}
	return nil
}

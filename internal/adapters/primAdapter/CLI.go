// Package primadapter stays for CLI interaction model.
package primadapter

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
	primports "github.com/Reddetk/crawler-cli/internal/ports/primPorts"
)

// Outcome reports how the crawl run finished
type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeInterrupted
	OutcomeTimeout
	OutcomeFailed
)

// ExitCode returns the process exit code for the outcome
func (o Outcome) ExitCode() int {
	switch o {
	case OutcomeSuccess:
		return 0
	case OutcomeInterrupted:
		return 130
	default:
		return 1
	}
}

type SeedProducer struct {
	WebParser primports.WebParser
	appCnf    *config.AppConfig
}

type CLI struct {
	urls     []string
	seedprod *SeedProducer
	log      logger.Logger
}

func NewCLI() *CLI {
	return &CLI{}
}

// Bind attaches the web parser, app config and logger to the CLI
func (cli *CLI) Bind(p primports.WebParser, appCnf *config.AppConfig, log logger.Logger) {
	cli.seedprod = &SeedProducer{WebParser: p, appCnf: appCnf}
	cli.log = log
}

// Run executes the crawl, persists the result and reports the outcome
func (cli *CLI) Run(ctx context.Context) Outcome {
	if cli.seedprod == nil {
		return OutcomeFailed
	}

	pages, err := cli.StartSeedProdusing(ctx)

	if werr := persist(cli.seedprod.appCnf.ResultPath, pages); werr != nil {
		cli.log.Error("write result failed", logger.Error(werr))
		return OutcomeFailed
	}

	outcome := classify(err)
	cli.report(outcome, err)
	return outcome
}

// StartSeedProdusing runs the crawl bounded by the overall app timeout
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

// ParseFlags fills configs and payload from console flags
func (cli *CLI) ParseFlags(cnf *config.AppConfig, srvCnf *config.ServiceConfig) (*config.AppConfig, *config.ServiceConfig, error) {
	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var errs []error

	timeout := fs.Duration("timeout", cnf.AppTimeout, "overall timeout for the whole crawl run, e.g. 2m, 90s, 1h30m")
	output := fs.String("output", cnf.ResultPath, "result output path")
	logPath := fs.String("log", cnf.LogCnf.OutputPaths[0], "additional logger output path")
	reqTimeout := fs.Duration("request-timeout", srvCnf.RequestTimeout, "timeout for a single HTTP request")
	depth := fs.Int("depth", srvCnf.Depth, "maximum link-following depth from each start URL (0 = only the start URL itself)")
	urlsStr := fs.String("urls", "", "comma-separated list of start URLs")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cnf, srvCnf, fmt.Errorf("failed to parse flags %w", err)
	}

	cnf.AppTimeout = *timeout
	srvCnf.RequestTimeout = *reqTimeout
	srvCnf.Depth = *depth

	if err := ensureParentDir(*output); err != nil {
		errs = append(errs, fmt.Errorf("output: %w", err))
	} else {
		cnf.ResultPath = *output
	}

	if *logPath != "" {
		if *logPath != "stdout" && *logPath != "stderr" {
			if err := ensureParentDir(*logPath); err != nil {
				errs = append(errs, fmt.Errorf("log: %w", err))
			} else {
				cnf.LogCnf.OutputPaths = []string{*logPath}
				cnf.LogCnf.ErrorOutputPaths = []string{"stderr"}
			}
		} else {
			cnf.LogCnf.OutputPaths = []string{*logPath}
			cnf.LogCnf.ErrorOutputPaths = []string{"stderr"}
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

func ensureParentDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", dir, err)
	}

	return nil
}

func classify(err error) Outcome {
	switch {
	case err == nil:
		return OutcomeSuccess
	case errors.Is(err, context.Canceled):
		return OutcomeInterrupted
	case errors.Is(err, context.DeadlineExceeded):
		return OutcomeTimeout
	default:
		return OutcomeFailed
	}
}

func (cli *CLI) report(outcome Outcome, err error) {
	path := cli.seedprod.appCnf.ResultPath
	switch outcome {
	case OutcomeSuccess:
		fmt.Fprintf(os.Stderr, "result saved to %s\n", path)
	case OutcomeInterrupted:
		cli.log.Info("crawl interrupted", logger.Error(err))
		fmt.Fprintln(os.Stderr, "interrupted: partial result saved to", path)
	case OutcomeTimeout:
		cli.log.Error("crawl failed", logger.Error(err))
		fmt.Fprintln(os.Stderr, "crawl timed out: partial result saved to", path)
	default:
		cli.log.Error("crawl failed", logger.Error(err))
		fmt.Fprintln(os.Stderr, "crawl failed: see log for details")
	}
}

func persist(path string, pages []*entity.Page) error {
	if pages == nil {
		pages = []*entity.Page{}
	}
	data, err := json.MarshalIndent(pages, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

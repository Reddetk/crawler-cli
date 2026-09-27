package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	primadapter "github.com/Reddetk/crawler-cli/internal/adapters/primAdapter"
	secadapter "github.com/Reddetk/crawler-cli/internal/adapters/secAdapter"
	"github.com/Reddetk/crawler-cli/internal/core"
	"github.com/Reddetk/crawler-cli/internal/core/entity"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli := primadapter.NewCLI()
	appCnf, srvCnf, warn := cli.ParseFlags(initConfigs())

	log, err := logger.NewZapLogger(*appCnf.LogCnf)
	if err != nil {
		panic("error of logger init")
	}
	if warn != nil {
		log.Warn("configuration warning:", logger.Error(warn))
	}

	crawler := core.NewCrawlerService(srvCnf, secadapter.NewHTTPObserver(log), log)
	cli.Bind(crawler, appCnf)

	pages, err := cli.StartSeedProdusing(ctx)
	if len(pages) > 0 {
		if werr := writeResult(appCnf.ResultPath, pages); werr != nil {
			log.Error("write result failed", logger.Error(werr))
			os.Exit(1)
		}
	}

	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "interrupted: partial result saved to", appCnf.ResultPath)
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, "crawl timed out: partial result saved to", appCnf.ResultPath)
		log.Error("crawl failed", logger.Error(err))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "result saved to %s\n", appCnf.ResultPath)
}

// writeResult stores the crawl result to a json file
func writeResult(path string, pages []*entity.Page) error {
	data, err := json.MarshalIndent(pages, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func initConfigs() (*config.AppConfig, *config.ServiceConfig) {
	return config.InitDefaultAppConfig(), config.InitDefaultServiceConfig()
}

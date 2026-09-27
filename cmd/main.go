package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	primadapter "github.com/Reddetk/crawler-cli/internal/adapters/primAdapter"
	secadapter "github.com/Reddetk/crawler-cli/internal/adapters/secAdapter"
	"github.com/Reddetk/crawler-cli/internal/core"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli := primadapter.NewCLI()
	appCnf, srvCnf, warn := cli.ParseFlags(initConfigs())
	if warn != nil {
		fmt.Fprintln(os.Stderr, warn)
		os.Exit(2)
	}

	log, err := logger.NewZapLogger(*appCnf.LogCnf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger init:", err)
		os.Exit(1)
	}

	crawler := core.NewCrawlerService(srvCnf, secadapter.NewHTTPObserver(log), log)
	cli.Bind(crawler, appCnf, log)

	os.Exit(cli.Run(ctx).ExitCode())
}

func initConfigs() (*config.AppConfig, *config.ServiceConfig) {
	return config.InitDefaultAppConfig(), config.InitDefaultServiceConfig()
}

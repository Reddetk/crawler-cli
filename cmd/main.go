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

	defAppCnf := config.InitDefaultAppConfig()
	defSrvCnf, err := config.InitDefaultServiceConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cli := primadapter.NewCLI()
	appCnf, srvCnf, warn := cli.ParseFlags(defAppCnf, defSrvCnf)
	if warn != nil {
		fmt.Fprintln(os.Stderr, warn)
		os.Exit(2)
	}

	log, err := logger.NewZapLogger(*appCnf.LogCnf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger init:", err)
		os.Exit(1)
	}

	crawler, err := core.NewCrawlerService(srvCnf, secadapter.NewHTTPObserver(log), log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "crawler service init:", err)
		os.Exit(1)
	}
	cli.Bind(crawler, appCnf, log)

	os.Exit(cli.Run(ctx).ExitCode())
}

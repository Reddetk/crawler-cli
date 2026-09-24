package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
)

// type app struct{}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logCnf := config.InitLoggerConfig()
	cnf := config.InitConfig(logCnf)

	runApp(ctx, cnf)

	<-signalChan

	fmt.Println("Received termination signal, shutting down...")
}

func runApp(ctx context.Context, cnf config.Config) {
	// logger initialization
	log, err := logger.NewZapLogger(cnf.LogCnf)
	if err != nil {
		panic("logger init error")
	}
	log.Info("Hi")
}

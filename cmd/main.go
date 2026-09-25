package main

import (
	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	primadapter "github.com/Reddetk/crawler-cli/internal/adapters/primAdapter"
)

func main() {
	cli := primadapter.NewCLI()

	appCnf, _, warn := cli.ParseFlags(initConfigs()) // TODO _
	log, err := logger.NewZapLogger(*appCnf.LogCnf)
	if err != nil {
		panic("error of logger init")
	}
	if warn != nil {
		log.Warn("configuration warning:", logger.Error(warn))
	}
}

func initConfigs() (*config.AppConfig, *config.ServiceConfig) {
	appcnf := config.InitDefaultAppConfig()
	srvcnf := config.InitDefaultServiceConfig()
	return appcnf, srvcnf
}

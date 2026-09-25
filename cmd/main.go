package main

import (
	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
	primadapter "github.com/Reddetk/crawler-cli/internal/adapters/primAdapter"
)

func main() {
	cli := primadapter.NewCLI()

	appCnf, srvCnf, warn := cli.ParseFlags(initConfigs())
	log, err := logger.NewZapLogger(*appCnf.LogCnf)
	if err != nil {
		panic("error of logger init")
	}
	if warn != nil {
		log.Warn("configuration warning:", logger.Error(warn))
	}

	primadapter.NewSeedProduser(l)
}

func initConfigs() (*config.AppConfig, *config.ServiceConfig) {
	appcnf := config.InitDefaultAppConfig()
	srvcnf := config.InitDefaultServiceConfig()
	return appcnf, srvcnf
}

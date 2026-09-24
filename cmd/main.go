package main

import (
	"github.com/Reddetk/crawler-cli/cmd/config"
	"github.com/Reddetk/crawler-cli/cmd/logger"
)


func main() {
	cnf, cnfErrors := configurateApp()

	log, err := logger.NewZapLogger(*cnf.LogCnf)
	if err != nil {
		panic(err)
	}

	if cnfErrors != nil {
		log.Warn("configuration errors:", logger.Error(cnfErrors))
	}

}


func configurateApp()(*config.Config, error){
	cnf, err := config.InitDefaultConfig()
	if err != nil {
		panic(err)
	}
	cnf, err  = cnf.ParseAppConfigFlags()
	return cnf, err
}

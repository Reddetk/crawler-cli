// Package config stays for acumulation configuration in one point
package config

import (
	"os"
	"strconv"
	"time"
)

const (
	DefaultRequepstTimeout time.Duration = time.Minute
	DefaultDepth           int           = 3
)

const (
	defaultMaxWorkers       int           = 5
	defaultAppTimeout       time.Duration = time.Minute * 2
	defaultResultPath       string        = "resources/result.json"
	defaultOutputPaths      string        = "resources/crawler.log"
	defaultErrorOutputPaths string        = "resources/crawler.log"
)

const (
	ENVNameMaxWorkers = "MAXWORKERS"
)

type AppConfig struct {
	AppTimeout time.Duration
	ResultPath string
	AppEnvs    appEnvs
	LogCnf     *LoggerConfig
}

type ServiceConfig struct {
	RequestTimeout time.Duration
	Depth          int
}

type appEnvs struct {
	maxWorkers int
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
}

func InitDefaultAppConfig() *AppConfig {
	return &AppConfig{
		AppEnvs:    parseEnv(),
		ResultPath: defaultResultPath,
		AppTimeout: defaultAppTimeout,
		LogCnf:     initDefaultLoggerConfig(),
	}
}

func InitDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		RequestTimeout: DefaultRequepstTimeout,
		Depth:          DefaultDepth,
	}
}

func initDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		OutputPaths:      []string{defaultOutputPaths},
		ErrorOutputPaths: []string{defaultErrorOutputPaths},
	}
}

// WithOutputPath stands for additional loggger output pathes
func (lCnf *LoggerConfig) WithOutputPath(paths ...string) *LoggerConfig {
	lCnf.OutputPaths = append(lCnf.OutputPaths, paths...)
	return lCnf
}

// WithErrOutputPath stands for additional loggger error output pathes
func (lCnf *LoggerConfig) WithErrOutputPath(paths ...string) *LoggerConfig {
	lCnf.ErrorOutputPaths = append(lCnf.ErrorOutputPaths, paths...)
	return lCnf
}

func parseEnv() appEnvs {
	return appEnvs{
		maxWorkers: getMaxWorkers(),
	}
}

func getMaxWorkers() int {
	plainmaxWorkers := os.Getenv(ENVNameMaxWorkers)
	maxWorkers, err := strconv.Atoi(plainmaxWorkers)
	if err != nil {
		maxWorkers = defaultMaxWorkers
	}
	return maxWorkers
}

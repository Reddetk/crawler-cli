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
	DefaultMaxWorkers       int           = 5
	DefaultAppTimeout       time.Duration = time.Minute * 2
	DefaultResultPath       string        = "resources/result.json"
	DefaultOutputPaths      string        = "resources/crawler.log"
	DefaultErrorOutputPaths string        = "resources/crawler.log"
)

const (
	ENVNameMaxWorkers = "MAXWORKERS"
)

type AppConfig struct {
	AppTimeout time.Duration
	ResultPath string
	LogCnf     *LoggerConfig
}

type ServiceConfig struct {
	RequestTimeout time.Duration
	AppEnvs        envs
	Depth          int
}

type envs struct {
	MaxWorkers int
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
}

func InitDefaultAppConfig() *AppConfig {
	return &AppConfig{
		ResultPath: DefaultResultPath,
		AppTimeout: DefaultAppTimeout,
		LogCnf:     initDefaultLoggerConfig(),
	}
}

func InitDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		AppEnvs:        parseEnv(),
		RequestTimeout: DefaultRequepstTimeout,
		Depth:          DefaultDepth,
	}
}

func initDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		OutputPaths:      []string{DefaultOutputPaths},
		ErrorOutputPaths: []string{DefaultErrorOutputPaths},
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

func parseEnv() envs {
	return envs{
		MaxWorkers: getMaxWorkers(),
	}
}

func getMaxWorkers() int {
	plainmaxWorkers := os.Getenv(ENVNameMaxWorkers)
	maxWorkers, err := strconv.Atoi(plainmaxWorkers)
	if err != nil {
		maxWorkers = DefaultMaxWorkers
	}
	return maxWorkers
}

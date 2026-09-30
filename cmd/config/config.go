// Package config stays for acumulation configuration in one point
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

const (
	MinWorkers = 1
	MaxWorkers = 10
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

func InitDefaultServiceConfig() (*ServiceConfig, error) {
	godotenv.Load()
	envs, err := parseEnv()
	return &ServiceConfig{
		AppEnvs:        envs,
		RequestTimeout: DefaultRequepstTimeout,
		Depth:          DefaultDepth,
	}, err
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

func parseEnv() (envs, error) {
	maxWorkers, err := getMaxWorkers()
	return envs{
		MaxWorkers: maxWorkers,
	}, err
}

func getMaxWorkers() (int, error) {
	raw := strings.TrimSpace(os.Getenv(ENVNameMaxWorkers))
	if raw == "" {
		return DefaultMaxWorkers, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not an integer: %w",
			ENVNameMaxWorkers, raw, err)
	}

	if err := ValidateMaxWorkers(n); err != nil {
		return 0, fmt.Errorf("%s: %w", ENVNameMaxWorkers, err)
	}

	return n, nil
}

func ValidateMaxWorkers(n int) error {
	if n < MinWorkers || n > MaxWorkers {
		return fmt.Errorf("invalid MAXWORKERS: got %d, want %d..%d", n, MinWorkers, MaxWorkers)
	}
	return nil
}

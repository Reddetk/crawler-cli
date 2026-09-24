// Package config stays for acumulation configuration in one point
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	defaultMaxWorkers int = 5
	defaultDepth int = 3
	defaultAppTimeout time.Duration = time.Minute*2
	defaultResultPath string = "resources/result.json"
	defaultOutputPaths      string = "resources/crawler.log"
	defaultErrorOutputPaths string = "resources/crawler.log"
)

const (
	ENVNameMaxWorkers = "MAXWORKERS"
)

type Config struct {
	AppTimeout time.Duration
	Depth int
	ResultPath string
	AppEnvs appEnvs 
	LogCnf *LoggerConfig
}

type appEnvs struct {
	maxWorkers int
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
}

func InitDefaultConfig() (*Config, error) {
	return &Config{
		AppEnvs: parseEnv(),
		Depth: defaultDepth,
		ResultPath: defaultResultPath,
		AppTimeout: defaultAppTimeout,
		LogCnf: initDefaultLoggerConfig(),
		}, nil
}


func initDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		OutputPaths: []string{defaultOutputPaths},
		ErrorOutputPaths: []string{defaultErrorOutputPaths},
	}
}

//ParseAppConfigFlags build config
//If val for config doesn't set, leave default
func (cnf *Config) ParseAppConfigFlags() (*Config, error) {

	fs := flag.NewFlagSet("crawler", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // подавить автоматический usage-вывод в stderr


	depth := fs.Int("depth", cnf.Depth, "crawl depth")
	timeout := fs.Duration("timeout", cnf.AppTimeout,
		"overall timeout for the whole crawl run, e.g. 2m, 90s, 1h30m. Format: Go time.Duration")
	output := fs.String("output", cnf.ResultPath, "result output path")
	logPath := fs.String("log", "", "additional logger output path")

	if err := fs.Parse(os.Args[1:]); err != nil {
		err := fmt.Errorf("failed to parse flags %w", err)
		return cnf, err
	}

	cnf.Depth = *depth
	cnf.AppTimeout = *timeout

	var errs []error

	if err := validatePath(*output); err != nil {
		errs = append(errs, fmt.Errorf("output: %w", err))
	} else {
		cnf.ResultPath = *output
	}

	if *logPath != "" {
		if err := validatePath(*logPath); err != nil {
			errs = append(errs, fmt.Errorf("log: %w", err))
		} else {
			cnf.LogCnf.WithOutputPath(*logPath)
		}
	}

	return cnf, errors.Join(errs...)
}

//WithOutputPath stands for additional loggger output pathes
func (lCnf *LoggerConfig) WithOutputPath(paths ...string) *LoggerConfig {
	lCnf.OutputPaths = append(lCnf.OutputPaths, paths...)
	return lCnf
}

//WithErrOutputPath stands for additional loggger error output pathes
func (lCnf *LoggerConfig) WithErrOutputPath(paths ...string) *LoggerConfig {
	lCnf.ErrorOutputPaths = append(lCnf.ErrorOutputPaths, paths...)
	return lCnf
}

//helpers
func validatePath(path string) error {
    _, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("path does not exist: %s", path)
        }
        return fmt.Errorf("cannot stat path %s: %w", path, err)
    }
    return nil
}


func parseEnv() (appEnvs){
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
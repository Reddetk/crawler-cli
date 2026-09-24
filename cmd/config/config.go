// Package config stays for acumulation configuration in one point
package config

type Config struct {
	LogCnf LoggerConfig
}

type LoggerConfig struct {
	OutputPaths      []string
	ErrorOutputPaths []string
}

func InitLoggerConfig() LoggerConfig {
	outpath := make([]string, 2)
	errpath := make([]string, 2)
	return LoggerConfig{
		OutputPaths:      outpath,
		ErrorOutputPaths: errpath,
	}
}

func (lCnf LoggerConfig) WithOutputPath(path ...string) LoggerConfig {
	lCnf.OutputPaths = append(lCnf.OutputPaths, path...)
	return lCnf
}

func (lCnf LoggerConfig) WithErrOutputPath(path ...string) LoggerConfig {
	lCnf.ErrorOutputPaths = append(lCnf.ErrorOutputPaths, path...)
	return lCnf
}

func InitConfig(logcnf LoggerConfig) Config {
	return Config{
		LogCnf: logcnf,
	}
}

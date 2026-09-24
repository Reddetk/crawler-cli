package logger

import (
	"github.com/Reddetk/crawler-cli/cmd/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Reexport zap.Field. for incapsulation

type Field = zap.Field

var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Bool     = zap.Bool
	Error    = zap.Error
	Duration = zap.Duration
	Any      = zap.Any
)

type zapLogger struct {
	l *zap.Logger
}

func (z *zapLogger) Debug(msg string, fields ...Field) { z.l.Debug(msg, fields...) }
func (z *zapLogger) Info(msg string, fields ...Field)  { z.l.Info(msg, fields...) }
func (z *zapLogger) Warn(msg string, fields ...Field)  { z.l.Warn(msg, fields...) }
func (z *zapLogger) Error(msg string, fields ...Field) { z.l.Error(msg, fields...) }
func (z *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{l: z.l.With(fields...)}
}

// NewZapLogger implement Logger with logger config
func NewZapLogger(logCnf config.LoggerConfig) (Logger, error) {
	zap.NewDevelopmentConfig()

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "crawler",
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"),
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      logCnf.OutputPaths,
		ErrorOutputPaths: logCnf.ErrorOutputPaths,
	}

	l, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &zapLogger{l: l}, err
}

// Package logger provides a wrapper
// Logger interface is used across all layers to avoid
// importing zap directly outside of cmd/ and infrastructure packages, and flex change to another framework without internal code cahnge
package logger

// Logger min interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

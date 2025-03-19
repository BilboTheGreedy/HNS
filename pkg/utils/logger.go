package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger with the specified log level
func InitLogger(level zerolog.Level) {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	log.Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

// GetRequestLogger returns a logger with request ID
func GetRequestLogger(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

// GetOperationLogger returns a logger with operation name
func GetOperationLogger(operation string) zerolog.Logger {
	return log.With().Str("operation", operation).Logger()
}

// LogLevel converts a string log level to a zerolog Level
func LogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
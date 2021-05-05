package util

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

const (
	logFormatJSON   = "json"
	logFormatPretty = "pretty"
)

func NewLogger(logContext string) zerolog.Logger {
	logFormat := os.Getenv("NEBRASKA_LOG_FORMAT")
	unknownFormat := false

	var writer io.Writer
	switch logFormat {
	case logFormatJSON:
		writer = os.Stderr
	case "", logFormatPretty:
		fallthrough
	default:
		writer = zerolog.ConsoleWriter{Out: os.Stderr}
		unknownFormat = true
	}

	logger := zerolog.New(writer).Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
			e.Str("context", logContext)
		}))

	if unknownFormat {
		logger.Debug().Str("logFormat", logFormat).Msg("Unknown format")
	}

	return logger
}

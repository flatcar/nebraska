package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

const (
	logFormatJSON   = "json"
	logFormatPretty = "pretty"
)

func New(logContext string) zerolog.Logger {
	logFormat := os.Getenv("NEBRASKA_LOG_FORMAT")
	unknownFormat := false

	var writer io.Writer
	switch logFormat {
	case logFormatJSON:
		writer = os.Stderr
	case "", logFormatPretty:
		fallthrough
	default:
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: zerolog.TimeFieldFormat,
		}
		unknownFormat = true
	}

	logger := zerolog.New(writer).With().Timestamp().Logger().Hook(
		zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
			e.Str("context", logContext)
		}))

	if unknownFormat {
		logger.Debug().Str("logFormat", logFormat).Msg("Unknown format")
	}

	return logger
}

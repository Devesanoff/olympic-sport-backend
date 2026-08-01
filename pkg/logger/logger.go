package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes global zerolog logger configuration.
func InitLogger(logLevel, appEnv string) zerolog.Logger {
	var level zerolog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	var logger zerolog.Logger
	if strings.ToLower(appEnv) == "development" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	log.Logger = logger
	return logger
}

// Get returns global zerolog logger instance.
func Get() *zerolog.Logger {
	return &log.Logger
}

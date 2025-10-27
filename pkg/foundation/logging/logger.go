package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var instance *Logger
var once sync.Once

type Logger struct {
	handler *CustomHandler
}

func Init() *Logger {
	once.Do(func() {
		logKeepDays, err := strconv.Atoi(os.Getenv("LOG_KEEP_DAYS"))

		if err != nil {
			panic(
				errors.New(
					fmt.Sprintf(
						"LOG_KEEP_DAYS is not set or invalid: %v",
						err,
					),
				),
			)
		}

		logDir := os.Getenv("LOG_DIR")

		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, 0755)

			if err != nil {
				panic(
					fmt.Sprintf(
						"failed to create log directory [%s]: %v",
						logDir,
						err,
					),
				)
			}
		}

		customHandler, err := NewCustomHandler(
			NewLevelPolicy(getLogLevels()),
			logDir,
			logKeepDays,
		)

		slog.SetDefault(slog.New(customHandler))

		instance = &Logger{
			handler: customHandler,
		}
	})

	return instance
}

func (l *Logger) Close() error {
	slog.Warn("Closing logger")

	if l.handler != nil {
		return l.handler.Close()
	}

	return nil
}

func getLogLevels() []slog.Level {
	logLevels := strings.Split(os.Getenv("LOG_LEVELS"), ",")

	var slogLevels []slog.Level

	if slices.Index(logLevels, "any") == -1 {
		for _, logLevel := range logLevels {
			if logLevel == "" {
				continue
			}

			switch logLevel {
			case "debug":
				slogLevels = append(slogLevels, slog.LevelDebug)
			case "info":
				slogLevels = append(slogLevels, slog.LevelInfo)
			case "warn":
				slogLevels = append(slogLevels, slog.LevelWarn)
			case "error":
				slogLevels = append(slogLevels, slog.LevelError)
			default:
				panic(fmt.Errorf("unknown log level: %s", logLevel))
			}
		}
	}

	return slogLevels
}

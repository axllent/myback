package logger

import (
	"os"

	"github.com/apsdehal/go-logger"
)

var (
	// Log global
	log *logger.Logger
	// Level - Log level
	Level string
	// LogToFile - log to file
	LogToFile string
	// ShowTimestamps - the sever invokes this
	ShowTimestamps = true
)

// Log handler
func Log() *logger.Logger {
	if log == nil {
		log = initLogger()
	}

	return log
}

func initLogger() *logger.Logger {
	var l *logger.Logger

	logLevel := logger.NoticeLevel

	if Level == "v" {
		logLevel = logger.WarningLevel
	} else if Level == "vv" {
		logLevel = logger.NoticeLevel
	} else if Level == "vvv" {
		logLevel = logger.InfoLevel
	} else if Level == "vvvv" {
		logLevel = logger.DebugLevel
	}

	l, _ = logger.New("proxy", 1, os.Stdout, logLevel)

	if ShowTimestamps {
		l.SetFormat("%{time} [%{level}] %{message}")
	} else {
		l.SetFormat("%{message}")
	}

	return l
}

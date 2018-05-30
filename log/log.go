// Package log provider wrapper of logger, filter lower level log, and add prefix for different level.
package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// log levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

// log prefix for different level.
const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

// Logger is a wrapper of log.Logger.
type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File
}

// New create a new logger according the log level, output file path and log flag.
func New(strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(file, "", flag)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}

	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile

	return logger, nil
}

// doPrintf do print according different level.
func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	logger.baseLogger.Output(3, fmt.Sprintf(format, a...))

	if level == fatalLevel {
		os.Exit(1)
	}
}

// Debug do a debug level log.
func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

// Release do a release level log.
func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

// Error do a error level log.
func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

// Fatal do a fatal level log and call os.Exit(1).
func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

// Close do close the logger.
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}
	logger.baseLogger = nil
	logger.baseFile = nil
}

// create default logger.
var gLogger, _ = New("debug", "", log.LstdFlags)

// Export external logger to replace internal logger.
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

// Debug is a debug level of log.
func Debug(format string, a ...interface{}) {
	gLogger.Debug(format, a...)
}

// Release is a release level of log.
func Release(format string, a ...interface{}) {
	gLogger.Release(format, a...)
}

// Error is a error level of log.
func Error(format string, a ...interface{}) {
	gLogger.Error(format, a...)
}

// Release is a release level of log.
func Fatal(format string, a ...interface{}) {
	gLogger.Fatal(format, a...)
}

// Close do close the internal logger.
func Close() {
	gLogger.Close()
}

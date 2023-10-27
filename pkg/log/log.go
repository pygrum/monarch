package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	logFile = "monarch.log"
	// LevelDebug logs all messages.
	LevelDebug uint16 = iota
	// LevelInfo logs only informational messages and any others with greater severity.
	LevelInfo
	// LevelWarn logs only warning messages and any others with greater severity.
	LevelWarn
	// LevelFatal logs only fatal messages.
	LevelFatal
	// ConsoleLogger creates a new console logger when used in NewLogger().
	ConsoleLogger
	// FileLogger creates a new file logger when used in NewLogger().
	FileLogger
)

// Logger interface declares methods used by both console loggers and file loggers.
// Console loggers display their output to the user, while file loggers
// write the output to a file. This functionality is useful for high volume and/or
// less urgent log messages, such as failed requests or other information.
type Logger interface {
	// Fatal is used when an application encounters a critical error, requiring
	// a shutdown.
	Fatal(format string, v ...interface{})
	// Warn is typically used to warn about errors that an application doesn't necessarily
	// have to shut down for.
	Warn(format string, v ...interface{})
	// Info is used for useful application notifications, such as an operation
	// completed successfully, or network operations such as received / sent requests and
	// responses.
	Info(format string, v ...interface{})
	// Debug is used for debug messages.
	Debug(format string, v ...interface{})
	// SetLogLevel determines which messages are logged based on their severity. For
	// example, a developer (or user) may choose to not log benign messages (debug, info)
	// to reduce noise.
	SetLogLevel(logLevel uint16) error
	// Close closes the logger and performs any necessary cleanup routines.
	Close() error
}

// Settings for loggers
type settings struct {
	// logLevel for the logger
	logLevel uint16
}

// Implementing the Logger interface, and is used for writing log messages to the console.
type consoleLogger struct {
	settings
}

// Implementing the Logger interface, and is used for writing log files.
type fileLogger struct {
	settings
	logFile *os.File
}

func NewLogger(loggerType uint16) (Logger, error) {
	var logger Logger
	switch loggerType {
	case ConsoleLogger:
		logger = &consoleLogger{
			settings: settings{
				logLevel: LevelInfo,
			},
		}
	case FileLogger:
		// create logfile to write to
		f, err := os.Create(filepath.Join(os.TempDir(), logFile))
		if err != nil {
			return nil, fmt.Errorf("failed to create log file: %v", err)
		}
		logger = &fileLogger{
			settings: settings{
				logLevel: LevelInfo,
			},
			logFile: f,
		}
	default:
		return nil, errors.New("invalid logger type specified")
	}
	return logger, nil
}

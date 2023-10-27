package log

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func t() string {
	return time.Now().Format(time.DateTime)
}

func (f *fileLogger) Fatal(format string, v ...interface{}) {
	// ignore returned error
	_, _ = fmt.Fprintln(f.logFile, t(), "[FATAL]", fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (f *fileLogger) Warn(format string, v ...interface{}) {
	if f.logLevel <= LevelWarn {
		_, _ = fmt.Fprintln(f.logFile, t(), "[WARN]", fmt.Sprintf(format, v...))
	}
}

func (f *fileLogger) Info(format string, v ...interface{}) {
	if f.logLevel <= LevelInfo {
		_, _ = fmt.Fprintln(f.logFile, t(), "[INFO]", fmt.Sprintf(format, v...))
	}
}

func (f *fileLogger) Debug(format string, v ...interface{}) {
	if f.logLevel <= LevelDebug {
		_, _ = fmt.Fprintln(f.logFile, t(), "[DEBUG]", fmt.Sprintf(format, v...))
	}
}

func (f *fileLogger) SetLogLevel(logLevel uint16) error {
	if logLevel < LevelDebug || logLevel > LevelFatal {
		return errors.New("invalid log level provided")
	}
	f.logLevel = LevelDebug
	return nil
}

func (f *fileLogger) Close() error {
	return f.logFile.Close()
}

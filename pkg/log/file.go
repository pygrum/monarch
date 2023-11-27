package log

import (
	"errors"
	"os"
)

func (f *fileLogger) Fatal(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// ignore returned error
	f.lrus.Fatalf(format, v...)
	os.Exit(1)
}

func (f *fileLogger) Error(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.logLevel <= LevelError {
		f.lrus.Errorf(format, v...)
	}
}

func (f *fileLogger) Warn(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.logLevel <= LevelWarn {
		f.lrus.Warnf(format, v...)
	}
}

func (f *fileLogger) Success(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.logLevel <= LevelInfo {
		f.lrus.Infof(format, v...)
	}
}

func (f *fileLogger) Info(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.logLevel <= LevelInfo {
		f.lrus.Infof(format, v...)
	}
}

func (f *fileLogger) Debug(format string, v ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.logLevel <= LevelDebug {
		f.lrus.Debugf(format, v...)
	}
}

func (f *fileLogger) SetLogLevel(logLevel uint16) error {
	if logLevel < LevelDebug || logLevel > LevelFatal {
		return errors.New("invalid log level provided")
	}
	f.logLevel = logLevel
	return nil
}

func (f *fileLogger) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.logFile.Close()
}

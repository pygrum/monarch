package log

import (
	"errors"
	"fmt"
	"os"
)

var (
	colorReset   = "\033[0m"
	colorFatal   = "\033[31m"
	colorInfo    = "\033[32m"
	colorWarning = "\033[33m"
	colorDebug   = "\033[34m"

	consoleFatalPrefix   = "[x]"
	consoleWarningPrefix = "[!]"
	consoleInfoPrefix    = "[+]"
	consoleDebugPrefix   = "[-]"
)

func (c *consoleLogger) Fatal(format string, v ...interface{}) {
	// ignore returned error
	fmt.Println(colorFatal+consoleFatalPrefix, fmt.Sprintf(format, v...), colorReset)
	os.Exit(1)
}

func (c *consoleLogger) Warn(format string, v ...interface{}) {
	if c.logLevel <= LevelWarn {
		fmt.Println(colorWarning+consoleWarningPrefix, fmt.Sprintf(format, v...),
			colorReset)
	}
}

func (c *consoleLogger) Info(format string, v ...interface{}) {
	if c.logLevel <= LevelInfo {
		fmt.Println(colorInfo+consoleInfoPrefix, fmt.Sprintf(format, v...), colorReset)
	}
}

func (c *consoleLogger) Debug(format string, v ...interface{}) {
	if c.logLevel <= LevelDebug {
		fmt.Println(colorDebug+consoleDebugPrefix, fmt.Sprintf(format, v...),
			colorReset)
	}
}

func (c *consoleLogger) SetLogLevel(logLevel uint16) error {
	if logLevel < LevelDebug || logLevel > LevelFatal {
		return errors.New("invalid log level provided")
	}
	c.logLevel = LevelDebug
	return nil
}

func (c *consoleLogger) Close() error {
	return nil
}

// DisableColor disables color output to the terminal for console loggers.
// Log levels can still be differentiated by their prefixes.
func DisableColor() {
	colorReset = ""
	colorFatal = ""
	colorInfo = ""
	colorWarning = ""
	colorDebug = ""
}

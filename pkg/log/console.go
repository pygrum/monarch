package log

import (
	"errors"
	"fmt"
	"os"
)

var (
	colorReset   = "\033[0m"
	colorError   = "\033[31m" // Red
	colorFatal   = "\033[37m" // Gray
	colorInfo    = "\033[34m" // Blue
	colorSuccess = "\033[32m"
	colorWarning = "\033[33m"
	colorDebug   = "\033[34m"

	consoleFatalPrefix   = "[x]"
	consoleErrorPrefix   = "[!]"
	consoleWarningPrefix = "[?]"
	consoleInfoPrefix    = "[*]"
	consoleSuccessPrefix = "[+]"
	consoleDebugPrefix   = "[-]"
)

func Print(args ...interface{}) {
	_, _ = app.Println(args...)
}

func (c *consoleLogger) Fatal(format string, v ...interface{}) {
	// ignore returned error
	fmt.Println(colorFatal+consoleFatalPrefix, fmt.Sprintf(format, v...), colorReset)
	os.Exit(1)
}

func (c *consoleLogger) Error(format string, v ...interface{}) {
	if c.logLevel <= LevelError {
		fmt.Println(colorError+consoleErrorPrefix, fmt.Sprintf(format, v...),
			colorReset)
	}
}

func (c *consoleLogger) Warn(format string, v ...interface{}) {
	if c.logLevel <= LevelWarn {
		fmt.Println(colorWarning+consoleWarningPrefix, fmt.Sprintf(format, v...),
			colorReset)
	}
}

func (c *consoleLogger) Success(format string, v ...interface{}) {
	if c.logLevel <= LevelSuccess {
		fmt.Println(colorSuccess+consoleSuccessPrefix, fmt.Sprintf(format, v...), colorReset)
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
	c.logLevel = logLevel
	return nil
}

func (c *consoleLogger) Close() error {
	return nil
}

func (c *transientLogger) Fatal(format string, v ...interface{}) {
	// ignore returned error
	fmt.Println(colorFatal + consoleFatalPrefix + " " + fmt.Sprintf(format, v...) + colorReset)
	os.Exit(1)
}

func (c *transientLogger) Error(format string, v ...interface{}) {
	if c.logLevel <= LevelError {
		_, _ = fmt.Fprintln(app.Stderr(), colorError+consoleErrorPrefix+" "+fmt.Sprintf(format, v...)+
			colorReset)
	}
}

func (c *transientLogger) Warn(format string, v ...interface{}) {
	if c.logLevel <= LevelWarn {
		_, _ = app.Println(colorWarning + consoleWarningPrefix + " " + fmt.Sprintf(format, v...) +
			colorReset)
	}
}

func (c *transientLogger) Success(format string, v ...interface{}) {
	if c.logLevel <= LevelSuccess {
		_, _ = app.Println(colorSuccess + consoleSuccessPrefix + " " + fmt.Sprintf(format, v...) + colorReset)
	}
}

func (c *transientLogger) Info(format string, v ...interface{}) {
	if c.logLevel <= LevelInfo {
		_, _ = app.Println(colorInfo + consoleInfoPrefix + " " + fmt.Sprintf(format, v...) + colorReset)
	}
}

func (c *transientLogger) Debug(format string, v ...interface{}) {
	if c.logLevel <= LevelDebug {
		_, _ = app.Println(colorDebug + consoleDebugPrefix + " " + fmt.Sprintf(format, v...) +
			colorReset)
	}
}

func (c *transientLogger) SetLogLevel(logLevel uint16) error {
	if logLevel < LevelDebug || logLevel > LevelFatal {
		return errors.New("invalid log level provided")
	}
	c.logLevel = logLevel
	return nil
}

func (c *transientLogger) Close() error {
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

package utils

import (
	"fmt"
	"log"
	"os"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

// Logger provides structured logging functionality with colored output.
type Logger struct {
	prefix        string
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
	warnLogger    *log.Logger
	successLogger *log.Logger
	useColor      bool
}

// NewLogger creates a new Logger instance with colored output support.
// Color output is automatically disabled if output is not a terminal.
func NewLogger(prefix string) *Logger {
	// Check if output is a terminal (enable colors only for TTY)
	useColor := isTerminal(os.Stdout)

	return &Logger{
		prefix:        prefix,
		infoLogger:    log.New(os.Stdout, "", log.LstdFlags),
		errorLogger:   log.New(os.Stderr, "", log.LstdFlags),
		debugLogger:   log.New(os.Stdout, "", log.LstdFlags),
		warnLogger:    log.New(os.Stdout, "", log.LstdFlags),
		successLogger: log.New(os.Stdout, "", log.LstdFlags),
		useColor:      useColor,
	}
}

// isTerminal checks if the file descriptor is a terminal.
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// formatMessage formats a log message with color and prefix.
func (l *Logger) formatMessage(level, color, format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	if l.useColor {
		return fmt.Sprintf("%s[%s]%s %s %s", color, level, colorReset, l.prefix, msg)
	}
	return fmt.Sprintf("[%s] %s %s", level, l.prefix, msg)
}

// Info logs an informational message in blue.
func (l *Logger) Info(format string, v ...interface{}) {
	msg := l.formatMessage("INFO", colorBlue, format, v...)
	l.infoLogger.Output(2, msg)
}

// Success logs a success message in green.
func (l *Logger) Success(format string, v ...interface{}) {
	msg := l.formatMessage("SUCCESS", colorGreen, format, v...)
	l.successLogger.Output(2, msg)
}

// Warn logs a warning message in yellow.
func (l *Logger) Warn(format string, v ...interface{}) {
	msg := l.formatMessage("WARN", colorYellow, format, v...)
	l.warnLogger.Output(2, msg)
}

// Error logs an error message in red.
func (l *Logger) Error(format string, v ...interface{}) {
	msg := l.formatMessage("ERROR", colorRed, format, v...)
	l.errorLogger.Output(2, msg)
}

// Debug logs a debug message in gray.
func (l *Logger) Debug(format string, v ...interface{}) {
	msg := l.formatMessage("DEBUG", colorGray, format, v...)
	l.debugLogger.Output(2, msg)
}

// Fatal logs a fatal error in red and exits the program.
func (l *Logger) Fatal(format string, v ...interface{}) {
	msg := l.formatMessage("FATAL", colorRed, format, v...)
	l.errorLogger.Output(2, msg)
	os.Exit(1)
}

// Step logs a step indicator for multi-step processes (in cyan).
// Usage: logger.Step(1, 5, "Connecting to database")
func (l *Logger) Step(current, total int, format string, v ...interface{}) {
	stepMsg := fmt.Sprintf("[%d/%d] ", current, total)
	msg := fmt.Sprintf(format, v...)
	if l.useColor {
		fullMsg := fmt.Sprintf("%s[STEP]%s %s %s%s%s", colorCyan, colorReset, l.prefix, colorCyan, stepMsg, colorReset) + msg
		l.infoLogger.Output(2, fullMsg)
	} else {
		fullMsg := fmt.Sprintf("[STEP] %s %s%s", l.prefix, stepMsg, msg)
		l.infoLogger.Output(2, fullMsg)
	}
}

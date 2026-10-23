package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Logger represents a logger
type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	level string
	file  *os.File
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warn:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
		error: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		level: "info",
	}
}

// SetLogFile adds a file to the logger output
func (l *Logger) SetLogFile(logPath string) error {
	if logPath == "" {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writers for each level
	if l.file != nil {
		_ = l.file.Close()
	}
	infoWriter := io.MultiWriter(os.Stdout, file)
	errorWriter := io.MultiWriter(os.Stderr, file)

	l.info.SetOutput(infoWriter)
	l.warn.SetOutput(infoWriter)
	l.error.SetOutput(errorWriter)
	l.file = file
	return nil
}

// SetLevel configures the minimum severity written by the logger.
func (l *Logger) SetLevel(level string) error {
	level = strings.ToLower(strings.TrimSpace(level))
	if level == "" {
		level = "info"
	}

	switch level {
	case "debug", "info", "warn", "error":
		l.level = level
		return nil
	default:
		return fmt.Errorf("unsupported log level %q", level)
	}
}

// Configure applies the configured log path and minimum severity.
func (l *Logger) Configure(logPath, level string) error {
	if err := l.SetLevel(level); err != nil {
		return err
	}
	if err := l.SetLogFile(logPath); err != nil {
		return err
	}
	return nil
}

// Close releases the configured log file.
func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level == "warn" || l.level == "error" {
		return
	}
	l.info.Printf(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level == "error" {
		return
	}
	l.warn.Printf(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.error.Printf(msg, args...)
}

// GetLogFilePath returns the standard log file path
func GetLogFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".mangahub", "logs", "server.log"), nil
}

// OpenLogFile opens the log file for reading
func OpenLogFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("log file does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return file, nil
}

// ReadLogLines reads log lines from a file with optional filtering
func ReadLogLines(file *os.File, level string, maxLines int) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if filterLogLevel(line, level) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return only the last maxLines
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	return lines, nil
}

// filterLogLevel checks if a log line matches the requested level
func filterLogLevel(line, level string) bool {
	if level == "" {
		return true
	}
	level = strings.ToUpper(level)
	tag := fmt.Sprintf("[%s]", level)
	return strings.Contains(line, tag)
}

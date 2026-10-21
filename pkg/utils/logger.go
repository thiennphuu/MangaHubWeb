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
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warn:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
		error: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

// SetLogFile adds a file to the logger output
func (l *Logger) SetLogFile(logPath string) {
	if logPath == "" {
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		NewLogger().Error("failed to create log directory for %s: %v", logPath, err)
		return
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		NewLogger().Error("failed to open log file %s: %v", logPath, err)
		return
	}

	// Create multi-writers for each level
	infoWriter := io.MultiWriter(os.Stdout, file)
	errorWriter := io.MultiWriter(os.Stderr, file)

	l.info.SetOutput(infoWriter)
	l.warn.SetOutput(infoWriter)
	l.error.SetOutput(errorWriter)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.info.Printf(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
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

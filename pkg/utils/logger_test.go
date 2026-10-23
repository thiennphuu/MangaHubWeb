package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggerConfigureWritesConfiguredFileAndHonorsLevel(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "logs", "server.log")
	logger := NewLogger()

	if err := logger.Configure(logPath, "error"); err != nil {
		t.Fatalf("configure logger: %v", err)
	}
	defer logger.Close()

	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "info message") || strings.Contains(content, "warn message") {
		t.Fatalf("log file contains messages below configured level: %q", content)
	}
	if !strings.Contains(content, "[ERROR]") || !strings.Contains(content, "error message") {
		t.Fatalf("log file does not contain error message: %q", content)
	}
}

func TestLoggerRejectsUnsupportedLevel(t *testing.T) {
	if err := NewLogger().SetLevel("trace"); err == nil {
		t.Fatal("expected unsupported log level error")
	}
}

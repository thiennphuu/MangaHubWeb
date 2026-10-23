package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigAppliesEnvironmentOverrides(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
app:
  jwt_secret: yaml-secret
http:
  host: 127.0.0.1
  port: 8080
  cors_origins:
    - http://localhost:3000
`)
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	t.Setenv("MANGAHUB_JWT_SECRET", "environment-secret")
	t.Setenv("MANGAHUB_HTTP_PORT", "8181")
	t.Setenv("MANGAHUB_CORS_ORIGINS", "https://web.example, http://localhost:3000")
	t.Setenv("MANGAHUB_LOG_LEVEL", "warn")

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.App.JWTSecret != "environment-secret" {
		t.Fatalf("expected environment JWT secret, got %q", config.App.JWTSecret)
	}
	if config.HTTP.Port != 8181 {
		t.Fatalf("expected environment HTTP port, got %d", config.HTTP.Port)
	}
	if len(config.HTTP.CORSOrigins) != 2 || config.HTTP.CORSOrigins[0] != "https://web.example" {
		t.Fatalf("unexpected CORS origins: %#v", config.HTTP.CORSOrigins)
	}
	if config.App.Logging.Level != "warn" {
		t.Fatalf("expected warning log level, got %q", config.App.Logging.Level)
	}
}

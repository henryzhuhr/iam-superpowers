package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create a temp YAML file
	content := `
database:
  host: testhost
  port: 5433
  name: testdb
  user: testuser
  password: testpass
  sslmode: disable
  maxConnections: 10
  idleTimeout: 120
redis:
  host: redishost
  port: 6380
  db: 1
  password: redispass
jwt:
  secret: test-secret
  accessTokenTTL: 300
  refreshTokenTTL: 86400
smtp:
  host: smtphost
  port: 587
  user: smtpuser
  password: smtppass
  from: test@example.com
  useTLS: true
server:
  port: 9090
  mode: test
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify values
	if cfg.Database.Host != "testhost" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "testhost")
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %d, want 5433", cfg.Database.Port)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("Database.Name = %q, want %q", cfg.Database.Name, "testdb")
	}
	if cfg.Redis.Host != "redishost" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "redishost")
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("JWT.Secret = %q, want %q", cfg.JWT.Secret, "test-secret")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Should not error when file not found — viper falls back to defaults/env
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load() should not error on missing file, got: %v", err)
	}
	// Defaults should be zero values
	if cfg.Server.Port != 0 {
		t.Errorf("Expected zero Port, got %d", cfg.Server.Port)
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigSearchesConfigSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	content := `{
	  "server": {
	    "host": "127.0.0.1",
	    "port": 50051,
	    "enable_reflection": true
	  },
	  "database": {
	    "host": "127.0.0.1",
	    "port": 5432,
	    "user": "test_user",
	    "password": "test_password",
	    "name": "test_db",
	    "ssl_mode": "disable",
	    "max_conns": 5,
	    "min_conns": 1
	  },
	  "security": {
	    "jwt_secret": "test-secret",
	    "token_ttl_minutes": 60
	  }
	}`
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Database.User != "test_user" {
		t.Fatalf("expected database user test_user, got %q", cfg.Database.User)
	}
	if cfg.Database.Password != "test_password" {
		t.Fatalf("expected database password test_password, got %q", cfg.Database.Password)
	}
}

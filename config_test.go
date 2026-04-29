package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "sshpass.config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestParseConfigFileAllowsExternalAuth(t *testing.T) {
	path := writeTempFile(t, "host: example.com\nusername: deploy\nport: 2222\n")

	cfg, err := parseConfigFile(path)
	if err != nil {
		t.Fatalf("parseConfigFile returned error: %v", err)
	}
	if cfg.Host != "example.com" || cfg.User != "deploy" || cfg.Port != "2222" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Password != "" || cfg.KeyPath != "" {
		t.Fatalf("config auth should be empty before CLI/env overrides: %+v", cfg)
	}
}

func TestLoadConfigOrPasswordFilePasswordOverridesConfig(t *testing.T) {
	path := writeTempFile(t, "host: example.com\nuser: root\npassword: from-file\nport: 2222\n")

	cfg, pass, err := loadConfigOrPasswordFile(path, "from-cli", true)
	if err != nil {
		t.Fatalf("loadConfigOrPasswordFile returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if pass != "from-cli" || cfg.Password != "from-cli" {
		t.Fatalf("password override failed: pass=%q config=%+v", pass, cfg)
	}
	if !cfg.StrictHostKey {
		t.Fatal("strict host key flag was not applied")
	}
}

func TestLoadConfigOrPasswordFileFallsBackToPasswordFile(t *testing.T) {
	path := writeTempFile(t, "secret\n")

	cfg, pass, err := loadConfigOrPasswordFile(path, "", false)
	if err != nil {
		t.Fatalf("loadConfigOrPasswordFile returned error: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected no config for password file, got %+v", cfg)
	}
	if pass != "secret" {
		t.Fatalf("password = %q, want %q", pass, "secret")
	}
}

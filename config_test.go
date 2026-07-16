package sshpass

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

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
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

	cfg, pass, err := LoadConfigOrPasswordFile(path, "from-cli", true)
	if err != nil {
		t.Fatalf("LoadConfigOrPasswordFile returned error: %v", err)
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

	cfg, pass, err := LoadConfigOrPasswordFile(path, "", false)
	if err != nil {
		t.Fatalf("LoadConfigOrPasswordFile returned error: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected no config for password file, got %+v", cfg)
	}
	if pass != "secret" {
		t.Fatalf("password = %q, want %q", pass, "secret")
	}
}

func TestParseConfigFileStrictHostKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"true", "host: h\nstrict_host_key: true\n", true},
		{"yes", "host: h\nstrict_host_key: yes\n", true},
		{"1", "host: h\nstrict_host_key: 1\n", true},
		{"on", "host: h\nstrict_host_key: on\n", true},
		{"false", "host: h\nstrict_host_key: false\n", false},
		{"no", "host: h\nstrict_host_key: no\n", false},
		{"not set", "host: h\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.input)
			cfg, err := LoadConfig(path)
			if err != nil {
				t.Fatalf("LoadConfig returned error: %v", err)
			}
			if cfg.StrictHostKey != tt.want {
				t.Errorf("StrictHostKey = %v, want %v", cfg.StrictHostKey, tt.want)
			}
		})
	}
}

func TestParseConfigFileTimeouts(t *testing.T) {
	path := writeTempFile(t, "host: h\ntimeout: 60\nconnect_timeout: 5\n")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Timeout != 60 {
		t.Errorf("Timeout = %d, want 60", cfg.Timeout)
	}
	if cfg.ConnectTimeout != 5 {
		t.Errorf("ConnectTimeout = %d, want 5", cfg.ConnectTimeout)
	}
}

func TestParseConfigFileMissingHost(t *testing.T) {
	path := writeTempFile(t, "user: root\n")
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
	if err == ErrNotConfigFile {
		t.Fatalf("expected 'missing host' error, got ErrNotConfigFile")
	}
}

func TestParseConfigFileNoRecognizedKeys(t *testing.T) {
	path := writeTempFile(t, "just a password string\n")
	_, err := LoadConfig(path)
	if err != ErrNotConfigFile {
		t.Fatalf("expected ErrNotConfigFile, got %v", err)
	}
}

func TestParseConfigFileEmptyFile(t *testing.T) {
	path := writeTempFile(t, "")
	_, err := LoadConfig(path)
	if err != ErrNotConfigFile {
		t.Fatalf("expected ErrNotConfigFile, got %v", err)
	}
}

func TestLoadConfigOrPasswordFileConfigMissingHost(t *testing.T) {
	// A file with recognized keys but missing host should be an error,
	// not silently treated as a password file.
	path := writeTempFile(t, "user: root\n")
	_, _, err := LoadConfigOrPasswordFile(path, "", false)
	if err == nil {
		t.Fatal("expected error for config file missing host")
	}
	if err == ErrNotConfigFile {
		t.Fatalf("expected 'missing host' error, got ErrNotConfigFile")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name           string
		timeout        int
		connectTimeout int
		wantConnect    int
	}{
		{"no timeout", 0, 10, 10},
		{"connect < timeout", 30, 5, 5},
		{"connect >= timeout", 10, 10, 9},
		{"connect much larger", 5, 100, 4},
		{"timeout 1 with connect 10", 1, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Timeout: tt.timeout, ConnectTimeout: tt.connectTimeout}
			cfg.Normalize()
			if cfg.ConnectTimeout != tt.wantConnect {
				t.Errorf("ConnectTimeout = %d, want %d", cfg.ConnectTimeout, tt.wantConnect)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	t.Run("src fields override dst", func(t *testing.T) {
		dst := NewConfig()
		dst.Host = "original"
		src := &Config{Host: "override", Password: "pass123"}
		dst.MergeConfig(src, nil)
		if dst.Host != "override" {
			t.Errorf("Host = %q, want %q", dst.Host, "override")
		}
		if dst.Password != "pass123" {
			t.Errorf("Password = %q, want %q", dst.Password, "pass123")
		}
	})

	t.Run("override takes priority over src", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{Host: "from-src", Password: "src-pass"}
		override := &Config{Host: "from-override", Password: "override-pass"}
		dst.MergeConfig(src, override)
		if dst.Host != "from-override" {
			t.Errorf("Host = %q, want %q", dst.Host, "from-override")
		}
		if dst.Password != "override-pass" {
			t.Errorf("Password = %q, want %q", dst.Password, "override-pass")
		}
	})

	t.Run("StrictHostKey merges from src", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{StrictHostKey: true}
		dst.MergeConfig(src, nil)
		if !dst.StrictHostKey {
			t.Error("StrictHostKey should be true from src")
		}
	})

	t.Run("StrictHostKey merges from override", func(t *testing.T) {
		dst := NewConfig()
		override := &Config{StrictHostKey: true}
		dst.MergeConfig(nil, override)
		if !dst.StrictHostKey {
			t.Error("StrictHostKey should be true from override")
		}
	})

	t.Run("empty override preserves src values", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{Host: "host1", Password: "pass1", Port: "2222"}
		override := &Config{}
		dst.MergeConfig(src, override)
		if dst.Host != "host1" {
			t.Errorf("Host = %q, want %q", dst.Host, "host1")
		}
		if dst.Password != "pass1" {
			t.Errorf("Password = %q, want %q", dst.Password, "pass1")
		}
		if dst.Port != "2222" {
			t.Errorf("Port = %q, want %q", dst.Port, "2222")
		}
	})

	t.Run("Timeout and ConnectTimeout 0 values are merged", func(t *testing.T) {
		dst := NewConfig() // Timeout=0, ConnectTimeout=10
		src := &Config{Host: "h", Timeout: 60, ConnectTimeout: 5}
		override := &Config{Timeout: 0, ConnectTimeout: 0}
		dst.MergeConfig(src, override)
		if dst.Timeout != 0 {
			t.Errorf("Timeout = %d, want 0 (override should apply)", dst.Timeout)
		}
		if dst.ConnectTimeout != 0 {
			t.Errorf("ConnectTimeout = %d, want 0 (override should apply)", dst.ConnectTimeout)
		}
	})

	t.Run("Timeout and ConnectTimeout -1 sentinel is skipped", func(t *testing.T) {
		dst := NewConfig() // Timeout=0, ConnectTimeout=10
		src := &Config{Host: "h", Timeout: 60, ConnectTimeout: 5}
		override := &Config{Timeout: -1, ConnectTimeout: -1}
		dst.MergeConfig(src, override)
		if dst.Timeout != 60 {
			t.Errorf("Timeout = %d, want 60 (sentinel -1 should be skipped)", dst.Timeout)
		}
		if dst.ConnectTimeout != 5 {
			t.Errorf("ConnectTimeout = %d, want 5 (sentinel -1 should be skipped)", dst.ConnectTimeout)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("missing host", func(t *testing.T) {
		cfg := &Config{Port: "22", Password: "pass"}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error for missing host")
		}
	})
	t.Run("invalid port", func(t *testing.T) {
		cfg := &Config{Host: "host", Port: "abc", Password: "pass"}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error for invalid port")
		}
	})
	t.Run("port out of range", func(t *testing.T) {
		cfg := &Config{Host: "host", Port: "99999", Password: "pass"}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error for port out of range")
		}
	})
	t.Run("missing auth", func(t *testing.T) {
		cfg := &Config{Host: "host", Port: "22"}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected error for missing auth")
		}
	})
	t.Run("valid with password", func(t *testing.T) {
		cfg := &Config{Host: "host", Port: "22", Password: "pass"}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("valid with key", func(t *testing.T) {
		cfg := &Config{Host: "host", Port: "22", KeyPath: "/key"}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestParseConfigFileProxyURL(t *testing.T) {
	path := writeTempFile(t, "host: h\nproxy: socks5://127.0.0.1:1080\n")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.ProxyURL != "socks5://127.0.0.1:1080" {
		t.Errorf("ProxyURL = %q, want %q", cfg.ProxyURL, "socks5://127.0.0.1:1080")
	}
}

func TestParseConfigFileProxyURLAlias(t *testing.T) {
	// "proxy_url" should be accepted as an alias for "proxy".
	path := writeTempFile(t, "host: h\nproxy_url: http://proxy:8080\n")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.ProxyURL != "http://proxy:8080" {
		t.Errorf("ProxyURL = %q, want %q", cfg.ProxyURL, "http://proxy:8080")
	}
}

func TestValidateAllowsEmptyProxyURL(t *testing.T) {
	// ProxyURL is optional; an empty value should pass validation.
	cfg := &Config{Host: "h", Port: "22", Password: "p"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate with empty ProxyURL should not error: %v", err)
	}
}

func TestValidateAllowsProxyURL(t *testing.T) {
	// A non-empty ProxyURL should not cause validation to fail (the URL is
	// only validated at dial time).
	cfg := &Config{Host: "h", Port: "22", Password: "p", ProxyURL: "socks5://x:1080"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate with ProxyURL should not error: %v", err)
	}
}

func TestMergeConfigProxyURL(t *testing.T) {
	t.Run("proxy from src overrides dst", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{Host: "h", ProxyURL: "http://proxy:8080"}
		dst.MergeConfig(src, nil)
		if dst.ProxyURL != "http://proxy:8080" {
			t.Errorf("ProxyURL = %q, want %q", dst.ProxyURL, "http://proxy:8080")
		}
	})
	t.Run("proxy from override takes priority", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{Host: "h", ProxyURL: "socks5://a:1080"}
		override := &Config{ProxyURL: "socks5://b:1080"}
		dst.MergeConfig(src, override)
		if dst.ProxyURL != "socks5://b:1080" {
			t.Errorf("ProxyURL = %q, want %q", dst.ProxyURL, "socks5://b:1080")
		}
	})
	t.Run("empty proxy preserves src", func(t *testing.T) {
		dst := NewConfig()
		src := &Config{Host: "h", ProxyURL: "http://p:8080"}
		dst.MergeConfig(src, &Config{})
		if dst.ProxyURL != "http://p:8080" {
			t.Errorf("ProxyURL = %q, want %q", dst.ProxyURL, "http://p:8080")
		}
	})
}

func TestParseBoolValue(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true}, {"yes", true}, {"1", true}, {"on", true},
		{"True", true}, {"YES", true}, {"ON", true},
		{"false", false}, {"no", false}, {"0", false}, {"off", false},
		{"maybe", false}, {"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseBoolValue(tt.input); got != tt.want {
				t.Errorf("parseBoolValue(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestApplyUserDefault(t *testing.T) {
	t.Run("sets root when empty", func(t *testing.T) {
		cfg := &Config{User: ""}
		cfg.ApplyUserDefault()
		if cfg.User != "root" {
			t.Errorf("User = %q, want root", cfg.User)
		}
	})
	t.Run("preserves existing user", func(t *testing.T) {
		cfg := &Config{User: "ubuntu"}
		cfg.ApplyUserDefault()
		if cfg.User != "ubuntu" {
			t.Errorf("User = %q, want ubuntu", cfg.User)
		}
	})
}

func TestSetUserHostFromArg(t *testing.T) {
	t.Run("parses user@host", func(t *testing.T) {
		cfg := NewConfig()
		cfg.SetUserHostFromArg("deploy@example.com")
		if cfg.User != "deploy" || cfg.Host != "example.com" {
			t.Errorf("User=%q Host=%q, want deploy/example.com", cfg.User, cfg.Host)
		}
	})
	t.Run("no at sign does nothing", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Host = "original"
		cfg.SetUserHostFromArg("justhost")
		if cfg.Host != "original" {
			t.Errorf("Host = %q, want original (unchanged)", cfg.Host)
		}
	})
	t.Run("ipv6 host", func(t *testing.T) {
		cfg := NewConfig()
		cfg.SetUserHostFromArg("root@[::1]:/tmp")
		if cfg.User != "root" || cfg.Host != "[::1]" {
			t.Errorf("User=%q Host=%q, want root/[::1]", cfg.User, cfg.Host)
		}
	})
}

func TestNormalizeEdgeCases(t *testing.T) {
	t.Run("timeout 0 does not adjust connect", func(t *testing.T) {
		cfg := &Config{Timeout: 0, ConnectTimeout: 10}
		cfg.Normalize()
		if cfg.ConnectTimeout != 10 {
			t.Errorf("ConnectTimeout = %d, want 10", cfg.ConnectTimeout)
		}
	})
	t.Run("connect < timeout unchanged", func(t *testing.T) {
		cfg := &Config{Timeout: 30, ConnectTimeout: 5}
		cfg.Normalize()
		if cfg.ConnectTimeout != 5 {
			t.Errorf("ConnectTimeout = %d, want 5", cfg.ConnectTimeout)
		}
	})
	t.Run("connect equals timeout clamps", func(t *testing.T) {
		cfg := &Config{Timeout: 10, ConnectTimeout: 10}
		cfg.Normalize()
		if cfg.ConnectTimeout != 9 {
			t.Errorf("ConnectTimeout = %d, want 9", cfg.ConnectTimeout)
		}
	})
	t.Run("timeout 1 connect 10 clamps to 1", func(t *testing.T) {
		cfg := &Config{Timeout: 1, ConnectTimeout: 10}
		cfg.Normalize()
		if cfg.ConnectTimeout != 1 {
			t.Errorf("ConnectTimeout = %d, want 1", cfg.ConnectTimeout)
		}
	})
}

func TestMergeFromNilIsNoop(t *testing.T) {
	dst := NewConfig()
	dst.Host = "original"
	dst.MergeFrom(nil)
	if dst.Host != "original" {
		t.Errorf("MergeFrom(nil) should not change dst, Host=%q", dst.Host)
	}
}

func TestNewConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	if cfg.User != "root" {
		t.Errorf("default User = %q, want root", cfg.User)
	}
	if cfg.Port != "22" {
		t.Errorf("default Port = %q, want 22", cfg.Port)
	}
	if cfg.ConnectTimeout != 10 {
		t.Errorf("default ConnectTimeout = %d, want 10", cfg.ConnectTimeout)
	}
	if cfg.Retries != 3 {
		t.Errorf("default Retries = %d, want 3", cfg.Retries)
	}
}

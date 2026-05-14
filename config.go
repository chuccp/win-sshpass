package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents SSH connection configuration
type Config struct {
	Host           string
	User           string
	Password       string
	Port           string
	KeyPath        string // private key file path
	StrictHostKey  bool   // whether to verify host key
	Timeout        int    // total operation deadline in seconds, 0 = no limit
	ConnectTimeout int    // TCP connection timeout in seconds
}

// newDefaultConfig creates a Config with default values
func newDefaultConfig() *Config {
	return &Config{
		User:           "root",
		Port:           "22",
		ConnectTimeout: 10,
	}
}

// applyUserDefault sets the user to "root" if empty
func applyUserDefault(cfg *Config) {
	if cfg.User == "" {
		cfg.User = "root"
	}
}

// normalize ensures config values are consistent
func (c *Config) normalize() {
	if c.Timeout > 0 && c.ConnectTimeout >= c.Timeout {
		c.ConnectTimeout = max(c.Timeout-1, 1)
	}
}

// setUserHostFromArg parses user@host:path format and sets config fields
func (c *Config) setUserHostFromArg(arg string) {
	user, host, _ := parseUserHostPath(arg)
	if user != "" && host != "" {
		c.User = user
		c.Host = host
	}
}

// validate checks that the config has required fields
func (c *Config) validate() error {
	if c.Host == "" {
		return fmt.Errorf("host address not specified")
	}
	if c.Password == "" && c.KeyPath == "" {
		return fmt.Errorf("no authentication method provided (password or key required)")
	}
	return nil
}

// mergeConfig merges non-empty fields from src into dst,
// then applies command-line overrides and user default.
func mergeConfig(dst, src, override *Config) {
	if src != nil {
		if src.Password != "" {
			dst.Password = src.Password
		}
		if src.KeyPath != "" {
			dst.KeyPath = src.KeyPath
		}
		if src.User != "" {
			dst.User = src.User
		}
		if src.Host != "" {
			dst.Host = src.Host
		}
		if src.Port != "" {
			dst.Port = src.Port
		}
		if src.Timeout > 0 {
			dst.Timeout = src.Timeout
		}
		if src.ConnectTimeout > 0 {
			dst.ConnectTimeout = src.ConnectTimeout
		}
	}
	if override != nil {
		if override.Password != "" {
			dst.Password = override.Password
		}
		if override.KeyPath != "" {
			dst.KeyPath = override.KeyPath
		}
		if override.Host != "" {
			dst.Host = override.Host
		}
		if override.User != "" {
			dst.User = override.User
		}
		if override.Port != "" && override.Port != "22" {
			dst.Port = override.Port
		}
		if override.Timeout > 0 {
			dst.Timeout = override.Timeout
		}
		if override.ConnectTimeout > 0 {
			dst.ConnectTimeout = override.ConnectTimeout
		}
	}
	applyUserDefault(dst)
	if dst.Timeout > 0 && dst.ConnectTimeout >= dst.Timeout {
		dst.ConnectTimeout = max(dst.Timeout-1, 1)
	}
}

// loadConfigOrPasswordFile treats filename as a config file first, falling back
// to a single-line password file when it is not a config.
func loadConfigOrPasswordFile(filename, password string, strictHostKey bool) (*Config, string, error) {
	pass := password

	config, err := parseConfigFile(filename)
	if err == nil {
		config.StrictHostKey = strictHostKey
		if pass != "" {
			config.Password = pass
		} else {
			pass = config.Password
		}
		return config, pass, nil
	}

	if pass != "" {
		if _, statErr := os.Stat(filename); statErr != nil {
			return nil, "", fmt.Errorf("failed to access config/password file: %w", statErr)
		}
		return nil, pass, nil
	}

	pass, err = readPasswordFile(filename)
	if err != nil {
		return nil, "", err
	}
	return nil, pass, nil
}

// parseConfigFile parses a config file (format: key: value)
func parseConfigFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := newDefaultConfig()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "host":
			config.Host = value
		case "user", "username":
			config.User = value
		case "password":
			config.Password = value
		case "port":
			config.Port = value
		case "key", "keypath":
			config.KeyPath = value
		case "timeout":
			if t, err := strconv.Atoi(value); err == nil && t >= 0 {
				config.Timeout = t
			}
		case "connect_timeout":
			if t, err := strconv.Atoi(value); err == nil && t >= 0 {
				config.ConnectTimeout = t
			}
		}
	}

	if config.Host == "" {
		return nil, fmt.Errorf("config file missing host")
	}

	return config, nil
}

// readPasswordFile reads password from a single-line password file
func readPasswordFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read password file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// getEnvPassword gets password from environment variable
func getEnvPassword() string {
	return os.Getenv("SSHPASS")
}

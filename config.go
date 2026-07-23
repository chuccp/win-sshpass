package sshpass

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents SSH connection configuration.
type Config struct {
	Host           string
	User           string
	Password       string
	Port           string
	KeyPath        string // private key file path
	StrictHostKey  bool   // whether to verify host key
	Timeout        int    // total operation deadline in seconds, 0 = no limit
	ConnectTimeout int    // TCP connection timeout in seconds
	Retries        int    // total connection attempts (default 1 = single attempt, no retry)
	ProxyURL       string // optional proxy URL (socks5://, socks5h://, socks4://, http://, https://)
	UseAgent       bool   // use ssh-agent for authentication
	AgentForward   bool   // enable ssh-agent forwarding to remote server
}

// NewConfig creates a Config with default values.
func NewConfig() *Config {
	return &Config{
		User:           "root",
		Port:           "22",
		ConnectTimeout: 10,
		Retries:        3,
	}
}

// ApplyUserDefault sets the user to "root" if empty.
func (c *Config) ApplyUserDefault() {
	if c.User == "" {
		c.User = "root"
	}
}

// Normalize ensures config values are consistent.
func (c *Config) Normalize() {
	if c.Timeout > 0 && c.ConnectTimeout >= c.Timeout {
		c.ConnectTimeout = max(c.Timeout-1, 1)
	}
}

// SetUserHostFromArg parses user@host:path format and sets config fields.
func (c *Config) SetUserHostFromArg(arg string) {
	user, host, _ := ParseUserHostPath(arg)
	if user != "" && host != "" {
		c.User = user
		c.Host = host
	}
}

// Validate checks that the config has required fields.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host address not specified")
	}
	if !isValidPort(c.Port) {
		return fmt.Errorf("invalid port number: %s (must be 1-65535)", c.Port)
	}
	if c.Password == "" && c.KeyPath == "" && !c.UseAgent {
		return fmt.Errorf("no authentication method provided (password, key, or ssh-agent required)")
	}
	return nil
}

// MergeFrom merges non-empty/non-sentinel fields from src into dst.
func (dst *Config) MergeFrom(src *Config) {
	if src == nil {
		return
	}
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
	if src.Timeout >= 0 {
		dst.Timeout = src.Timeout
	}
	if src.ConnectTimeout >= 0 {
		dst.ConnectTimeout = src.ConnectTimeout
	}
	if src.Retries >= 0 {
		dst.Retries = src.Retries
	}
	if src.StrictHostKey {
		dst.StrictHostKey = true
	}
	if src.ProxyURL != "" {
		dst.ProxyURL = src.ProxyURL
	}
	if src.UseAgent {
		dst.UseAgent = true
	}
	if src.AgentForward {
		dst.AgentForward = true
	}
}

// MergeConfig merges non-empty fields from src into dst, then applies overrides.
func (dst *Config) MergeConfig(src, override *Config) {
	dst.MergeFrom(src)
	dst.MergeFrom(override)
}

// LoadConfigOrPasswordFile treats filename as a config file first, falling back
// to a single-line password file when it is not a config. password is the
// already-known password (e.g. from -p); it takes priority over the file.
// strictHostKey, when true, forces StrictHostKey on the resulting config.
func LoadConfigOrPasswordFile(filename, password string, strictHostKey bool) (*Config, string, error) {
	pass := password

	config, err := LoadConfig(filename)
	if err == nil {
		// CLI -k flag overrides config file, but config file value is preserved if -k not set
		if strictHostKey {
			config.StrictHostKey = true
		}
		if pass != "" {
			config.Password = pass
		} else {
			pass = config.Password
		}
		return config, pass, nil
	}

	// If the file is a config file but has a real error (e.g. missing host),
	// report it instead of silently falling back to password file.
	if err != ErrNotConfigFile {
		return nil, "", err
	}

	// Not a config file — fall back to treating it as a password file.
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

// ErrNotConfigFile indicates the file does not contain recognized config keys.
var ErrNotConfigFile = fmt.Errorf("not a config file")

// LoadConfig parses a config file (format: key: value).
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := NewConfig()
	hasKeys := false

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
			hasKeys = true
		case "user", "username":
			config.User = value
			hasKeys = true
		case "password":
			config.Password = value
			hasKeys = true
		case "port":
			config.Port = value
			hasKeys = true
		case "key", "keypath":
			config.KeyPath = value
			hasKeys = true
		case "timeout":
			if t, err := strconv.Atoi(value); err == nil && t >= 0 {
				config.Timeout = t
			}
			hasKeys = true
		case "connect_timeout":
			if t, err := strconv.Atoi(value); err == nil && t >= 0 {
				config.ConnectTimeout = t
			}
			hasKeys = true
		case "retry", "retries":
			if t, err := strconv.Atoi(value); err == nil && t >= 0 {
				config.Retries = t
			}
			hasKeys = true
		case "strict_host_key":
			config.StrictHostKey = parseBoolValue(value)
			hasKeys = true
		case "proxy", "proxy_url":
			config.ProxyURL = value
			hasKeys = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if !hasKeys {
		return nil, ErrNotConfigFile
	}

	if config.Host == "" {
		return nil, fmt.Errorf("config file missing host")
	}

	return config, nil
}

// parseBoolValue parses a boolean value from a config file
func parseBoolValue(s string) bool {
	switch strings.ToLower(s) {
	case "true", "yes", "1", "on":
		return true
	default:
		return false
	}
}

// readPasswordFile reads password from a single-line password file
func readPasswordFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read password file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// GetEnvPassword returns the password from the SSHPASS environment variable.
func GetEnvPassword() string {
	return os.Getenv("SSHPASS")
}

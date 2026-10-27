package main

import (
	"os"
	"strings"
	"testing"
	"time"

	sshpass "github.com/chuccp/win-sshpass"
)

func TestApplyParsedSSHConfigPreservesConfigAndAppliesOverrides(t *testing.T) {
	config := &sshpass.Config{
		Host:       "configured.example",
		User:       "deploy",
		Port:       "22",
		KeyPath:    "/old/key",
		Timeout:    60,
		Background: false,
	}
	parsed, cmd := sshpass.ParseSSHArgs([]string{
		"ssh", "--bg", "-p", "2222", "-i", "/new/key", "root@target.example", "sleep 2",
	})

	applyParsedSSHConfig(config, parsed)

	if config.Host != "target.example" {
		t.Errorf("Host = %q, want target.example", config.Host)
	}
	if config.User != "root" {
		t.Errorf("User = %q, want root", config.User)
	}
	if config.Port != "2222" {
		t.Errorf("Port = %q, want 2222", config.Port)
	}
	if config.KeyPath != "/new/key" {
		t.Errorf("KeyPath = %q, want /new/key", config.KeyPath)
	}
	if !config.Background {
		t.Error("Background = false, want true")
	}
	if config.Timeout != 60 {
		t.Errorf("Timeout = %d, want existing value 60", config.Timeout)
	}
	if cmd != "sleep 2" {
		t.Errorf("command = %q, want sleep 2", cmd)
	}
}

func TestApplyParsedSSHConfigNilValues(t *testing.T) {
	applyParsedSSHConfig(nil, &sshpass.Config{})
	applyParsedSSHConfig(&sshpass.Config{}, nil)
}

func TestResolveConfigCommand(t *testing.T) {
	t.Run("positional command does not replace host", func(t *testing.T) {
		config := &sshpass.Config{Host: "configured.example", User: "deploy"}
		command := resolveConfigCommand(config, []string{"whoami", "--short"})
		if command != "whoami --short" {
			t.Errorf("command = %q, want whoami --short", command)
		}
		if config.Host != "configured.example" || config.User != "deploy" {
			t.Errorf("connection target changed: %+v", config)
		}
	})

	t.Run("explicit ssh args override target", func(t *testing.T) {
		config := &sshpass.Config{Host: "configured.example", User: "deploy"}
		command := resolveConfigCommand(config, []string{"ssh", "-p", "2222", "root@target.example", "uname", "-a"})
		if command != "uname -a" {
			t.Errorf("command = %q, want uname -a", command)
		}
		if config.Host != "target.example" || config.User != "root" || config.Port != "2222" {
			t.Errorf("target overrides not applied: %+v", config)
		}
	})
}

func TestParseForwardSpec(t *testing.T) {
	tests := []struct {
		name       string
		spec       string
		wantListen string
		wantTarget string
		wantError  bool
	}{
		{
			name:       "port host hostport",
			spec:       "8080:db.internal:3306",
			wantListen: "127.0.0.1:8080",
			wantTarget: "db.internal:3306",
		},
		{
			name:       "bind port host hostport",
			spec:       "0.0.0.0:8080:db.internal:3306",
			wantListen: "0.0.0.0:8080",
			wantTarget: "db.internal:3306",
		},
		{name: "too few fields", spec: "8080:db", wantError: true},
		{name: "too many fields", spec: "a:b:c:d:e", wantError: true},
		{name: "empty", spec: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listen, target, err := parseForwardSpec(tt.spec)
			if (err != nil) != tt.wantError {
				t.Fatalf("parseForwardSpec(%q) error = %v, wantError %v", tt.spec, err, tt.wantError)
			}
			if listen != tt.wantListen {
				t.Errorf("listen = %q, want %q", listen, tt.wantListen)
			}
			if target != tt.wantTarget {
				t.Errorf("target = %q, want %q", target, tt.wantTarget)
			}
		})
	}
}

func TestRepeatableFlag(t *testing.T) {
	var flags repeatableFlag
	if err := flags.Set("first"); err != nil {
		t.Fatal(err)
	}
	if err := flags.Set("second"); err != nil {
		t.Fatal(err)
	}
	if got := flags.String(); got != "first,second" {
		t.Errorf("String() = %q, want first,second", got)
	}
	if len(flags) != 2 || flags[0] != "first" || flags[1] != "second" {
		t.Errorf("collected values = %v, want [first second]", flags)
	}
}

func TestUpdateSummary(t *testing.T) {
	tests := []struct {
		name      string
		result    *sshpass.UpdateResult
		checkOnly bool
		want      string
	}{
		{
			name:   "updated",
			result: &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", Updated: true, TargetPath: "/bin/win-sshpass"},
			want:   "Updated win-sshpass v1.0.0 -> v1.1.0 (/bin/win-sshpass)",
		},
		{
			name:   "reinstalled",
			result: &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0", Updated: true, TargetPath: "/bin/win-sshpass"},
			want:   "Reinstalled win-sshpass v1.0.0 (/bin/win-sshpass)",
		},
		{
			name:      "check with update",
			result:    &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", UpdateAvailable: true, ReleaseURL: "https://example.test/release"},
			checkOnly: true,
			want:      "A newer version is available: v1.0.0 -> v1.1.0",
		},
		{
			name:      "check up to date",
			result:    &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0"},
			checkOnly: true,
			want:      "win-sshpass v1.0.0 is up to date",
		},
		{
			name:   "already latest",
			result: &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0"},
			want:   "win-sshpass v1.0.0 is already the latest version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateSummary(tt.result, tt.checkOnly)
			if !strings.Contains(got, tt.want) {
				t.Errorf("updateSummary() = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

func TestCurrentUserFallback(t *testing.T) {
	t.Setenv("USER", "primary")
	t.Setenv("USERNAME", "fallback")
	if got := currentUser(); got != "primary" {
		t.Errorf("currentUser() = %q, want primary", got)
	}

	t.Setenv("USER", "")
	if got := currentUser(); got != "fallback" {
		t.Errorf("currentUser() = %q, want fallback", got)
	}

	t.Setenv("USERNAME", "")
	if got := currentUser(); got != "user" {
		t.Errorf("currentUser() = %q, want user", got)
	}
}

func TestDefaultKeyComment(t *testing.T) {
	t.Setenv("USER", "tester")
	t.Setenv("USERNAME", "")
	comment := defaultKeyComment()
	if !strings.HasPrefix(comment, "tester") {
		t.Errorf("defaultKeyComment() = %q, want tester prefix", comment)
	}
}

func TestUpdateSummaryDoesNotMutateResult(t *testing.T) {
	result := &sshpass.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", Updated: true, TargetPath: "/tmp/win-sshpass"}
	before := *result
	_ = updateSummary(result, false)
	after := *result
	if before != after {
		t.Errorf("updateSummary mutated result: before=%+v after=%+v", before, after)
	}
}

func TestJSONStateAccessors(t *testing.T) {
	previous := jsonState
	t.Cleanup(func() { jsonState = previous })

	jsonState.enabled = false
	jsonState.host = ""
	jsonState.command = ""
	if jsonEnabled() {
		t.Fatal("jsonEnabled() = true, want false")
	}
	jsonSetHost("root@example.com")
	jsonSetCommand("uname -a")
	if jsonEnabled() {
		t.Fatal("jsonEnabled() changed unexpectedly")
	}

	jsonState.enabled = true
	if !jsonEnabled() {
		t.Fatal("jsonEnabled() = false, want true")
	}
	if jsonState.host != "root@example.com" || jsonState.command != "uname -a" {
		t.Errorf("JSON state = host %q command %q", jsonState.host, jsonState.command)
	}
}

func TestJSONDuration(t *testing.T) {
	previous := jsonState
	t.Cleanup(func() { jsonState = previous })

	jsonInit()
	jsonState.startTime = time.Now().Add(-1500 * time.Millisecond)
	if got := jsonDurationMs(); got < 1400 || got > 2500 {
		t.Errorf("jsonDurationMs() = %d, want about 1500", got)
	}
}

func TestCurrentHostname(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Skipf("os.Hostname unavailable: %v", err)
	}
	if got := currentHostname(); got != host {
		t.Errorf("currentHostname() = %q, want %q", got, host)
	}
}

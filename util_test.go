package sshpass

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testExitStatusError struct {
	code int
}

func (e testExitStatusError) Error() string {
	return "remote command failed"
}

func (e testExitStatusError) ExitStatus() int {
	return e.code
}

func TestExitCodeFromError(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", testExitStatusError{code: 7})

	code, ok := ExitCodeFromError(err)
	if !ok {
		t.Fatal("expected exit status error to be detected")
	}
	if code != 7 {
		t.Fatalf("code = %d, want 7", code)
	}
}

func TestExitCodeFromErrorNotMatched(t *testing.T) {
	code, ok := ExitCodeFromError(fmt.Errorf("plain error"))
	if ok {
		t.Fatal("should not match plain error")
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
}

func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"12a", false},
		{" 123", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isAllDigits(tt.input); got != tt.want {
				t.Errorf("isAllDigits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"22", true},
		{"1", true},
		{"65535", true},
		{"8080", true},
		{"0", false},
		{"65536", false},
		{"-1", false},
		{"abc", false},
		{"", false},
		{"22a", false},
		{"12.34", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isValidPort(tt.input); got != tt.want {
				t.Errorf("isValidPort(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsWindowsLocalPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"C:/Users/file.txt", true},
		{"D:\\data\\file.txt", true},
		{"c:/path", true},
		{"/tmp/file", false},
		{"C:", false},         // too short
		{"1:/invalid", false}, // digit drive letter
		{"", false},
		{"ab", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isWindowsLocalPath(tt.input); got != tt.want {
				t.Errorf("isWindowsLocalPath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanRemotePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"double slash prefix", "//tmp/file", "/tmp/file"},
		{"normal path", "/home/user/file.txt", "/home/user/file.txt"},
		{"single slash", "/tmp/file", "/tmp/file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CleanRemotePath(tt.input)
			if err != nil {
				t.Fatalf("CleanRemotePath(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("CleanRemotePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanRemotePathWindowsError(t *testing.T) {
	// A Windows-style local path (Git Bash conversion) should return an error
	// instead of terminating the process.
	_, err := CleanRemotePath("C:/Users/someone/tmp/file")
	if err == nil {
		t.Fatal("expected error for Windows-looking remote path")
	}
}

func TestParseUserHostPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUser string
		wantHost string
		wantPath string
	}{
		{"simple", "root@host:/tmp/file", "root", "host", "/tmp/file"},
		{"no path", "root@host", "root", "host", ""},
		{"ipv6", "root@[::1]:/tmp", "root", "[::1]", "/tmp"},
		{"ipv6 no path", "root@[2001:db8::1]", "root", "[2001:db8::1]", ""},
		{"no at sign", "host:path", "", "", ""},
		{"empty", "", "", "", ""},
		{"@ at start", "@host:path", "", "", ""},
		{"ipv6 with path no trailing", "user@[::1]:/var/log", "user", "[::1]", "/var/log"},
		{"user host with colon in path", "root@host:/a:b", "root", "host", "/a:b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, host, p := ParseUserHostPath(tt.input)
			if user != tt.wantUser {
				t.Errorf("user = %q, want %q", user, tt.wantUser)
			}
			if host != tt.wantHost {
				t.Errorf("host = %q, want %q", host, tt.wantHost)
			}
			if p != tt.wantPath {
				t.Errorf("path = %q, want %q", p, tt.wantPath)
			}
		})
	}
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"multiple", []string{"a", "b", "c"}, "a b c"},
		{"single", []string{"only"}, "only"},
		{"empty", []string{}, ""},
		{"with flags", []string{"-la", "/tmp"}, "-la /tmp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinArgs(tt.args); got != tt.want {
				t.Errorf("JoinArgs(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestSetupOperationTimeoutFires(t *testing.T) {
	var triggered atomic.Bool
	closeFn := func() { triggered.Store(true) }

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	reset, stop := setupOperationTimeout(logger, closeFn, 1)
	_ = reset
	defer stop()

	// Wait for the timer to fire (1 second + margin).
	time.Sleep(1500 * time.Millisecond)

	if !triggered.Load() {
		t.Fatal("expected timeout to trigger closeFn")
	}
	if !strings.Contains(buf.String(), "operation timed out") {
		t.Errorf("expected timeout notice in log output, got %q", buf.String())
	}
}

func TestSetupOperationTimeoutNoTimeout(t *testing.T) {
	// timeout <= 0: no timer, stop is a no-op, reset is nil.
	var triggered atomic.Bool
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reset, stop := setupOperationTimeout(logger, func() { triggered.Store(true) }, 0)
	defer stop()
	if reset != nil {
		t.Errorf("reset should be nil when timeout <= 0, got non-nil")
	}
	// Wait briefly; should not fire.
	time.Sleep(200 * time.Millisecond)
	if triggered.Load() {
		t.Fatal("closeFn should not fire when timeout <= 0")
	}
}

func TestSetupOperationTimeoutReset(t *testing.T) {
	var triggered atomic.Bool
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reset, stop := setupOperationTimeout(logger, func() { triggered.Store(true) }, 1)
	defer stop()

	// Reset multiple times within the window to keep it alive.
	for i := 0; i < 3; i++ {
		reset()
		time.Sleep(600 * time.Millisecond)
	}
	if triggered.Load() {
		t.Fatal("closeFn should not fire while being reset")
	}
}

func TestOnInterruptStopReleasesGoroutine(t *testing.T) {
	var called atomic.Bool
	stop := onInterrupt(func() { called.Store(true) })

	// Stopping should unregister the handler without calling cleanup.
	stop()

	// Give the goroutine time to exit.
	time.Sleep(100 * time.Millisecond)
	if called.Load() {
		t.Fatal("cleanup should not be called on stop")
	}
}

func TestSplitPaths(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"single path", "/tmp/file.txt", []string{"/tmp/file.txt"}, false},
		{"comma separated", "/a.txt,/b.txt", []string{"/a.txt", "/b.txt"}, false},
		{"space separated simple", "a.txt b.txt", []string{"a.txt", "b.txt"}, false},
		{"space separated with slash errors", "/a.txt /b.txt", nil, true},
		{"comma with spaces trimmed", " a.txt , b.txt ", []string{"a.txt", "b.txt"}, false},
		{"comma trailing empty", "a.txt,,b.txt", []string{"a.txt", "b.txt"}, false},
		{"empty string", "", []string{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitPaths(tt.input, "local")
			if (err != nil) != tt.wantErr {
				t.Fatalf("SplitPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Fatalf("SplitPaths() = %v, want %v", got, tt.want)
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("SplitPaths()[%d] = %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

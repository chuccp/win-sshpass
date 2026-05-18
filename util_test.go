package main

import (
	"fmt"
	"testing"
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

	code, ok := exitCodeFromError(err)
	if !ok {
		t.Fatal("expected exit status error to be detected")
	}
	if code != 7 {
		t.Fatalf("code = %d, want 7", code)
	}
}

func TestExitCodeFromErrorNotMatched(t *testing.T) {
	code, ok := exitCodeFromError(fmt.Errorf("plain error"))
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

func TestIsWindowsLocalPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"C:/Users/file.txt", true},
		{"D:\\data\\file.txt", true},
		{"c:/path", true},
		{"/tmp/file", false},
		{"C:", false},       // too short
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
			if got := cleanRemotePath(tt.input); got != tt.want {
				t.Errorf("cleanRemotePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseUserHostPath(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantUser   string
		wantHost   string
		wantPath   string
	}{
		{"simple", "root@host:/tmp/file", "root", "host", "/tmp/file"},
		{"no path", "root@host", "root", "host", ""},
		{"ipv6", "root@[::1]:/tmp", "root", "[::1]", "/tmp"},
		{"ipv6 no path", "root@[2001:db8::1]", "root", "[2001:db8::1]", ""},
		{"no at sign", "host:path", "", "", ""},
		{"empty", "", "", "", ""},
		{"@ at start", "@host:path", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, host, p := parseUserHostPath(tt.input)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitPaths(tt.input, "local")
			if (err != nil) != tt.wantErr {
				t.Fatalf("splitPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Fatalf("splitPaths() = %v, want %v", got, tt.want)
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("splitPaths()[%d] = %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestIsClosedConnError(t *testing.T) {
	if !isClosedConnError(fmt.Errorf("use of closed network connection")) {
		t.Error("should match closed connection error")
	}
	if isClosedConnError(fmt.Errorf("something else")) {
		t.Error("should not match other errors")
	}
}

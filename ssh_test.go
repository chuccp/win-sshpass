package sshpass

import (
	"net"
	"strings"
	"testing"
)

// dialAddress constructs the network address for dialing, stripping IPv6
// brackets from the host if present. This mirrors the logic in dial().
func dialAddress(host, port string) string {
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	return net.JoinHostPort(host, port)
}

func TestDialAddressIPv6(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{"IPv6 with brackets", "[::1]", "22", "[::1]:22"},
		{"IPv6 without brackets", "::1", "22", "[::1]:22"},
		{"IPv6 full", "[2001:db8::1]", "2222", "[2001:db8::1]:2222"},
		{"IPv4", "192.168.1.1", "22", "192.168.1.1:22"},
		{"hostname", "example.com", "22", "example.com:22"},
		{"localhost", "localhost", "2222", "localhost:2222"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dialAddress(tt.host, tt.port)
			if got != tt.want {
				t.Errorf("dialAddress(%q, %q) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
			// The result must be round-trippable through SplitHostPort.
			h, p, err := net.SplitHostPort(got)
			if err != nil {
				t.Fatalf("net.SplitHostPort(%q) error: %v", got, err)
			}
			if p != tt.port {
				t.Errorf("port = %q, want %q", p, tt.port)
			}
			_ = h
		})
	}
}

// TestDialAddressFromParseUserHostPath verifies the end-to-end flow:
// ParseUserHostPath returns a bracketed IPv6 host, and dialAddress correctly
// strips the brackets before passing to net.JoinHostPort.
func TestDialAddressFromParseUserHostPath(t *testing.T) {
	user, host, remotePath := ParseUserHostPath("root@[::1]:/tmp/file")
	if user != "root" {
		t.Fatalf("user = %q, want root", user)
	}
	if host != "[::1]" {
		t.Fatalf("host = %q, want [::1]", host)
	}
	if remotePath != "/tmp/file" {
		t.Fatalf("remotePath = %q, want /tmp/file", remotePath)
	}

	// This is what dial() does with the parsed host.
	addr := dialAddress(host, "22")
	if addr != "[::1]:22" {
		t.Fatalf("dialAddress produced %q, want [::1]:22", addr)
	}

	// Verify the address is valid for net.Dial.
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error: %v", addr, err)
	}
}

// TestDialAddressFromParseSSHArgs verifies that ParseSSHArgs + dialAddress
// produces a valid address for IPv6 hosts.
func TestDialAddressFromParseSSHArgs(t *testing.T) {
	cfg, _ := ParseSSHArgs([]string{"ssh", "root@[::1]"})
	if cfg.Host != "[::1]" {
		t.Fatalf("cfg.Host = %q, want [::1]", cfg.Host)
	}

	addr := dialAddress(cfg.Host, "22")
	if addr != "[::1]:22" {
		t.Fatalf("dialAddress produced %q, want [::1]:22", addr)
	}
}

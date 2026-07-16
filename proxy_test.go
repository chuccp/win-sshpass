package sshpass

import "testing"

func TestProxyDialUnsupportedScheme(t *testing.T) {
	_, err := proxyDial("ftp://proxy:21", "example.com:22", 5)
	if err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestProxyDialInvalidURL(t *testing.T) {
	// url.Parse is lenient with "://" missing; this should still parse but
	// yield an empty scheme, which we reject.
	_, err := proxyDial("not a url", "example.com:22", 5)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestProxyDialSchemeRouting(t *testing.T) {
	// These will fail at the network layer (no real proxy), but should fail
	// with a connection error, not a scheme-routing error — proving the
	// correct dialer was selected.
	cases := []struct {
		name string
		url  string
	}{
		{"socks5", "socks5://127.0.0.1:1"},
		{"socks5h", "socks5h://127.0.0.1:1"},
		{"socks4", "socks4://127.0.0.1:1"},
		{"http", "http://127.0.0.1:1"},
		{"https", "https://127.0.0.1:1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := proxyDial(tc.url, "example.com:22", 1)
			if err == nil {
				t.Fatal("expected connection error (no real proxy)")
			}
			// The error must come from the dialer, not from scheme routing.
			if msg := err.Error(); contains(msg, "unsupported proxy scheme") {
				t.Errorf("scheme was not routed correctly: %s", msg)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

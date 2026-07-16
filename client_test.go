package sshpass

import "testing"

func TestNewClientNilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestClientCloseIdempotent(t *testing.T) {
	// Construct a Client without dialing (sshClient is nil).
	// Close should be safe to call multiple times without panic.
	c := &Client{}

	if err := c.Close(); err != nil {
		t.Errorf("first Close error = %v, want nil", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("second Close error = %v, want nil (idempotent)", err)
	}
}

func TestClientConfigAccessor(t *testing.T) {
	cfg := NewConfig()
	c := &Client{config: cfg}
	if c.Config() != cfg {
		t.Error("Config() did not return the same config pointer")
	}
}

func TestClientTimedOutDefaultFalse(t *testing.T) {
	c := &Client{}
	if c.TimedOut() {
		t.Error("TimedOut should be false for a fresh client")
	}
}

func TestClientSSHClientAccessor(t *testing.T) {
	// Without a real connection, SSHClient() returns nil.
	c := &Client{}
	if c.SSHClient() != nil {
		t.Error("SSHClient() should be nil for unconnected client")
	}
}

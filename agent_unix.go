//go:build !windows

package sshpass

import (
	"fmt"
	"io"
	"net"
	"os"
)

// agentDial connects to the local ssh-agent via SSH_AUTH_SOCK (Unix domain
// socket). On Unix-like systems this is the standard transport.
func agentDial() (io.ReadWriteCloser, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set; is ssh-agent running?")
	}
	return net.Dial("unix", socket)
}

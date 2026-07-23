//go:build windows

package sshpass

import (
	"fmt"
	"io"
	"os"
)

// agentDial connects to the Windows OpenSSH Authentication Agent via its
// named pipe (\\.\pipe\openssh-ssh-agent). The pipe is opened as a regular
// file handle, which Go's os.OpenFile supports on Windows.
func agentDial() (io.ReadWriteCloser, error) {
	const pipePath = `\\.\pipe\openssh-ssh-agent`
	f, err := os.OpenFile(pipePath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ssh-agent (named pipe %s): %w\nhint: ensure the OpenSSH Authentication Agent service is running (Get-Service ssh-agent)", pipePath, err)
	}
	return f, nil
}

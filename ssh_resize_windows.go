//go:build windows

package sshpass

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// watchTerminalResize monitors terminal resize on Windows by polling.
// fd is the terminal file descriptor whose size is polled and reported to the
// remote session.
func watchTerminalResize(session *ssh.Session, done <-chan struct{}, fd int) {
	go func() {
		var lastCols, lastRows int
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				cols, rows := getTerminalSize(fd)
				if cols != lastCols || rows != lastRows {
					lastCols, lastRows = cols, rows
					sendWindowChange(session, fd)
				}
			}
		}
	}()
}

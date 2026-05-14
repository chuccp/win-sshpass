//go:build windows

package main

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// watchTerminalResize monitors terminal resize on Windows by polling
func watchTerminalResize(session *ssh.Session, done <-chan struct{}) {
	go func() {
		var lastCols, lastRows int
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				cols, rows := getTerminalSize()
				if cols != lastCols || rows != lastRows {
					lastCols, lastRows = cols, rows
					sendWindowChange(session)
				}
			}
		}
	}()
}

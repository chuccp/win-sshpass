//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// watchTerminalResize monitors terminal resize on Unix using SIGWINCH
func watchTerminalResize(session *ssh.Session, done <-chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		defer signal.Stop(sigChan)
		for {
			select {
			case <-done:
				return
			case <-sigChan:
				sendWindowChange(session)
			}
		}
	}()
}

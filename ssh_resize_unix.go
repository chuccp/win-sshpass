//go:build !windows

package sshpass

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// watchTerminalResize monitors terminal resize on Unix using SIGWINCH.
// fd is the terminal file descriptor whose size changes are reported to the
// remote session.
func watchTerminalResize(session *ssh.Session, done <-chan struct{}, fd int) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		defer signal.Stop(sigChan)
		for {
			select {
			case <-done:
				return
			case <-sigChan:
				sendWindowChange(session, fd)
			}
		}
	}()
}

//go:build !windows

package sshpass

import (
	"fmt"
	"os"
)

// replaceExecutable moves newBinary over target, preserving the permissions of
// the file it replaces.
//
// On Unix the running executable can be replaced directly: rename(2) swaps the
// directory entry atomically and this process keeps running from the old inode
// until it exits, so no temporary copy of the old binary is needed.
func replaceExecutable(target, newBinary string) error {
	mode := os.FileMode(0o755)
	if fi, err := os.Stat(target); err == nil {
		mode = fi.Mode().Perm()
	}
	if err := os.Chmod(newBinary, mode); err != nil {
		return fmt.Errorf("cannot set permissions on the new binary: %w", err)
	}
	if err := os.Rename(newBinary, target); err != nil {
		return fmt.Errorf("cannot replace %s: %w%s", target, err, privilegedInstallHint(target))
	}
	return nil
}

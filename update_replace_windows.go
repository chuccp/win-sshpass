//go:build windows

package sshpass

import (
	"fmt"
	"os"
	"time"
)

// oldSuffix is appended to the target to park the previous binary while the
// new one takes its place.
const oldSuffix = ".old"

// replaceExecutable moves newBinary over target.
//
// Windows refuses to overwrite a running executable, but it does allow the
// image file to be renamed. The current binary is therefore moved to
// "<target>.old" and the new one takes its place; the parked copy stays locked
// until this process exits, so it is deleted at the start of the next update
// (deleting it now is attempted anyway and ignored when it fails).
func replaceExecutable(target, newBinary string) error {
	old := target + oldSuffix

	// A copy parked by an earlier update is no longer mapped by any process.
	// If a win-sshpass is running elsewhere it is still in use, in which case
	// the rename below fails with a clear error instead of deleting it.
	_ = os.Remove(old)

	if err := renameRetrying(target, old); err != nil {
		return fmt.Errorf("cannot move the current binary aside: %w%s", err, privilegedInstallHint(target))
	}
	if err := renameRetrying(newBinary, target); err != nil {
		// Put the original back: never leave the install directory without a
		// working binary.
		if restoreErr := renameRetrying(old, target); restoreErr != nil {
			return fmt.Errorf("cannot install the new binary (%v) and cannot restore the original (%v); it is still available as %s", err, restoreErr, old)
		}
		return fmt.Errorf("cannot install the new binary: %w", err)
	}

	_ = os.Remove(old)
	return nil
}

// renameRetrying renames src to dst, retrying briefly. Windows returns a
// sharing violation when an anti-virus scanner or the indexer has the freshly
// written file open for a moment, and that clears up on its own.
func renameRetrying(src, dst string) error {
	var err error
	for attempt := 0; attempt < 4; attempt++ {
		if err = os.Rename(src, dst); err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

//go:build !windows

package main

// cliFileSelector is a no-op FileSelector for Linux and macOS.
// rz/sz file transfers fall back to reading a path from stdin instead of
// showing a GUI file dialog.
type cliFileSelector struct{}

// OpenFile returns empty with no error, causing the rz/sz handler to prompt
// for a file path on stdin.
func (cliFileSelector) OpenFile() (string, error) {
	return "", nil
}

// SaveFile returns empty with no error, causing the rz/sz handler to use the
// remote file's base name and prompt on stdin if needed.
func (cliFileSelector) SaveFile(defaultName string) (string, error) {
	return "", nil
}

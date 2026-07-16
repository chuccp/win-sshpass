//go:build windows

package main

import (
	"errors"

	"github.com/ncruces/zenity"
)

// cliFileSelector implements sshpass.FileSelector using zenity dialogs.
// On Windows, rz/sz file transfers pop up a native file open/save dialog.
type cliFileSelector struct{}

func (cliFileSelector) OpenFile() (string, error) {
	path, err := zenity.SelectFile(zenity.Title("Select file to upload"))
	if errors.Is(err, zenity.ErrCanceled) {
		return "", nil
	}
	return path, err
}

func (cliFileSelector) SaveFile(defaultName string) (string, error) {
	path, err := zenity.SelectFileSave(
		zenity.Title("Save downloaded file"),
		zenity.Filename(defaultName),
		zenity.ConfirmOverwrite(),
	)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", nil
	}
	return path, err
}

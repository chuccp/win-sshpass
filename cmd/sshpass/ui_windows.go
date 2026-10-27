//go:build windows

package main

import (
	"errors"

	"github.com/ncruces/zenity"
)

// cliFileSelector implements sshpass.FileSelector using zenity dialogs.
// On Windows, rz/sz file transfers pop up a native file open/save dialog.
type cliFileSelector struct{}

var (
	zenitySelectFile     = zenity.SelectFile
	zenitySelectFileSave = zenity.SelectFileSave
)

func (cliFileSelector) OpenFile() (string, error) {
	path, err := zenitySelectFile(zenity.Title("Select file to upload"))
	if errors.Is(err, zenity.ErrCanceled) {
		return "", nil
	}
	return path, err
}

func (cliFileSelector) SaveFile(defaultName string) (string, error) {
	path, err := zenitySelectFileSave(
		zenity.Title("Save downloaded file"),
		zenity.Filename(defaultName),
		zenity.ConfirmOverwrite(),
	)
	if errors.Is(err, zenity.ErrCanceled) {
		return "", nil
	}
	return path, err
}

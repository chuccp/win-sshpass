//go:build darwin

package main

import (
	"errors"

	"github.com/ncruces/zenity"
)

// cliFileSelector implements sshpass.FileSelector using zenity dialogs.
// On macOS, zenity uses osascript to show native Finder file dialogs.
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

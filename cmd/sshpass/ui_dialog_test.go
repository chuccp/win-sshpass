//go:build windows || darwin

package main

import (
	"errors"
	"testing"

	"github.com/ncruces/zenity"
)

func TestCLIFileSelectorOpenFileInterception(t *testing.T) {
	original := zenitySelectFile
	t.Cleanup(func() { zenitySelectFile = original })

	tests := []struct {
		name      string
		dialog    func(...zenity.Option) (string, error)
		wantPath  string
		wantError bool
	}{
		{
			name: "selected",
			dialog: func(options ...zenity.Option) (string, error) {
				if len(options) != 1 {
					t.Errorf("SelectFile options = %d, want 1", len(options))
				}
				return "C:/tmp/upload.txt", nil
			},
			wantPath: "C:/tmp/upload.txt",
		},
		{
			name: "canceled",
			dialog: func(...zenity.Option) (string, error) {
				return "", zenity.ErrCanceled
			},
		},
		{
			name: "dialog error",
			dialog: func(...zenity.Option) (string, error) {
				return "", errors.New("dialog failed")
			},
			wantError: true,
		},
	}

	selector := cliFileSelector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zenitySelectFile = tt.dialog
			path, err := selector.OpenFile()
			if (err != nil) != tt.wantError {
				t.Fatalf("OpenFile() error = %v, wantError %v", err, tt.wantError)
			}
			if path != tt.wantPath {
				t.Errorf("OpenFile() path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestCLIFileSelectorSaveFileInterception(t *testing.T) {
	original := zenitySelectFileSave
	t.Cleanup(func() { zenitySelectFileSave = original })

	tests := []struct {
		name      string
		dialog    func(...zenity.Option) (string, error)
		wantPath  string
		wantError bool
	}{
		{
			name: "selected",
			dialog: func(options ...zenity.Option) (string, error) {
				if len(options) != 3 {
					t.Errorf("SelectFileSave options = %d, want title, filename, and overwrite confirmation", len(options))
				}
				return "C:/tmp/download.txt", nil
			},
			wantPath: "C:/tmp/download.txt",
		},
		{
			name: "canceled",
			dialog: func(...zenity.Option) (string, error) {
				return "", zenity.ErrCanceled
			},
		},
		{
			name: "dialog error",
			dialog: func(...zenity.Option) (string, error) {
				return "", errors.New("dialog failed")
			},
			wantError: true,
		},
	}

	selector := cliFileSelector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zenitySelectFileSave = tt.dialog
			path, err := selector.SaveFile("remote.txt")
			if (err != nil) != tt.wantError {
				t.Fatalf("SaveFile() error = %v, wantError %v", err, tt.wantError)
			}
			if path != tt.wantPath {
				t.Errorf("SaveFile() path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
)

// timeoutReader wraps an io.Reader and calls reset on each successful Read.
// This keeps the operation timeout alive as long as data flows from the remote.
type timeoutReader struct {
	r     io.Reader
	reset func()
}

func (tr *timeoutReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 && tr.reset != nil {
		tr.reset()
	}
	return n, err
}

// timeoutWriter wraps an io.Writer and calls reset on each successful Write.
// This keeps the operation timeout alive as long as data flows to the remote.
type timeoutWriter struct {
	w     io.Writer
	reset func()
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	n, err := tw.w.Write(p)
	if n > 0 && tw.reset != nil {
		tw.reset()
	}
	return n, err
}


// uploadFile uploads a file or directory to the remote server.
// resetTimeout is called on each data transfer to keep the operation timeout alive.
func uploadFile(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func()) error {
	// check if local path is a file or directory
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to get local file info: %w", err)
	}

	if localInfo.IsDir() {
		return uploadDirectory(sftpClient, localPath, remotePath, resetTimeout)
	}
	return uploadSingleFile(sftpClient, localPath, remotePath, resetTimeout)
}

// uploadSingleFile uploads a single file
func uploadSingleFile(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func()) error {
	// check if remote path is a directory
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteFileInfo.IsDir() {
		remotePath = joinRemotePath(remotePath, localBaseName(localPath))
	} else if err != nil && strings.HasSuffix(remotePath, "/") {
		// path doesn't exist but has trailing slash: treat as directory target
		remotePath = joinRemotePath(remotePath, localBaseName(localPath))
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// get file size
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// ensure remote directory exists
	remoteDir := remoteDirName(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// create progress bar
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("Uploading %s", localBaseName(localPath))),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionUseANSICodes(true),
	)

	// wrap remote file writer to reset timeout on each data chunk
	remoteWriter := &timeoutWriter{w: remoteFile, reset: resetTimeout}
	_, err = io.Copy(remoteWriter, io.TeeReader(localFile, bar))
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// uploadDirectory uploads an entire directory
func uploadDirectory(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func()) error {
	// get local directory base name
	localBase := localBaseName(localPath)

	// ensure remote directory exists
	remoteDir := joinRemotePath(remotePath, localBase)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// walk local directory
	return filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// calculate relative path
		relPath, err := filepath.Rel(localPath, filePath)
		if err != nil {
			return err
		}

		// convert Windows relative path to Unix style
		relPath = toSlash(relPath)

		// remote full path
		remoteFullPath := joinRemotePath(remoteDir, relPath)

		if info.IsDir() {
			// create remote directory
			return sftpClient.MkdirAll(remoteFullPath)
		}

		// upload file
		return uploadSingleFile(sftpClient, filePath, remoteFullPath, resetTimeout)
	})
}

// downloadFile downloads a file or directory from the remote server.
// resetTimeout is called on each data transfer to keep the operation timeout alive.
func downloadFile(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func()) error {
	// check if remote path is a file or directory
	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}

	if remoteInfo.IsDir() {
		return downloadDirectory(sftpClient, remotePath, localPath, resetTimeout)
	}
	return downloadSingleFile(sftpClient, remotePath, localPath, resetTimeout)
}

// downloadSingleFile downloads a single file
func downloadSingleFile(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func()) error {
	// check if local path is a directory
	localFileInfo, err := os.Stat(localPath)
	if err == nil && localFileInfo.IsDir() {
		localPath = joinLocalPath(localPath, remoteBaseName(remotePath))
	}

	// ensure local directory exists
	localDir := localDirName(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// get file size
	fileInfo, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// create progress bar
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", remoteBaseName(remotePath))),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionUseANSICodes(true),
	)

	// wrap remote file reader to reset timeout on each data chunk
	remoteReader := &timeoutReader{r: remoteFile, reset: resetTimeout}
	_, err = io.Copy(localFile, io.TeeReader(remoteReader, bar))
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// flush data to disk for reliability
	if err := localFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// downloadDirectory downloads an entire directory
func downloadDirectory(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func()) error {
	// get remote directory name (trim trailing / to avoid remoteBaseName returning empty string)
	remoteBase := remoteBaseName(strings.TrimSuffix(remotePath, "/"))

	// ensure local directory exists
	localDir := joinLocalPath(localPath, remoteBase)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// ensure remote path ends with / for relative path calculation
	remotePath = strings.TrimSuffix(remotePath, "/") + "/"

	// walk remote directory
	walker := sftpClient.Walk(remotePath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		remoteFilePath := walker.Path()

		// calculate relative path (remove remote base path)
		relPath := strings.TrimPrefix(remoteFilePath, remotePath)
		if relPath == "" {
			continue
		}

		// local full path
		localFullPath := joinLocalPath(localDir, relPath)

		info := walker.Stat()
		if info.IsDir() {
			// create local directory
			if err := os.MkdirAll(localFullPath, 0755); err != nil {
				return err
			}
		} else {
			// download file
			if err := downloadSingleFile(sftpClient, remoteFilePath, localFullPath, resetTimeout); err != nil {
				return err
			}
		}
	}

	return nil
}

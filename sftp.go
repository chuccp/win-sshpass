package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
)

// ensureRemoteDir creates a remote directory if it doesn't exist
func ensureRemoteDir(client *ssh.Client, path string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()
	return sftpClient.MkdirAll(path)
}

// uploadFile uploads a file or directory to the remote server
func uploadFile(client *ssh.Client, localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// check if local path is a file or directory
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to get local file info: %w", err)
	}

	if localInfo.IsDir() {
		return uploadDirectory(sftpClient, localPath, remotePath)
	}
	return uploadSingleFile(sftpClient, localPath, remotePath)
}

// uploadSingleFile uploads a single file
func uploadSingleFile(sftpClient *sftp.Client, localPath, remotePath string) error {
	// check if remote path is a directory
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteFileInfo.IsDir() {
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

	_, err = io.Copy(remoteFile, io.TeeReader(localFile, bar))
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// uploadDirectory uploads an entire directory
func uploadDirectory(sftpClient *sftp.Client, localPath, remotePath string) error {
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
		return uploadSingleFile(sftpClient, filePath, remoteFullPath)
	})
}

// downloadFile downloads a file or directory from the remote server
func downloadFile(client *ssh.Client, remotePath, localPath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// check if remote path is a file or directory
	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}

	if remoteInfo.IsDir() {
		return downloadDirectory(sftpClient, remotePath, localPath)
	}
	return downloadSingleFile(sftpClient, remotePath, localPath)
}

// downloadSingleFile downloads a single file
func downloadSingleFile(sftpClient *sftp.Client, remotePath, localPath string) error {
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

	_, err = io.Copy(localFile, io.TeeReader(remoteFile, bar))
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// downloadDirectory downloads an entire directory
func downloadDirectory(sftpClient *sftp.Client, remotePath, localPath string) error {
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
			if err := downloadSingleFile(sftpClient, remoteFilePath, localFullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

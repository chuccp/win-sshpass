package sshpass

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// newHash returns a hash.Hash for the given algorithm name.
// Supported: md5, sha1, sha256, sha512 (case-insensitive).
func newHash(algo string) (hash.Hash, error) {
	switch strings.ToLower(algo) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unknown hash algorithm: %s (supported: md5, sha1, sha256, sha512)", algo)
	}
}

// HashFile computes the hex-encoded hash of a local file using the given
// algorithm. algo must be one of: md5, sha1, sha256, sha512.
// The file is read in a streaming manner to support large files.
func HashFile(localPath string, algo string) (string, error) {
	h, err := newHash(algo)
	if err != nil {
		return "", err
	}

	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyFile checks whether the local file's hash matches the expected hex
// string. Returns true when the hashes match (case-insensitive comparison).
// algo must be one of: md5, sha1, sha256, sha512.
func VerifyFile(localPath string, algo string, expected string) (bool, error) {
	actual, err := HashFile(localPath, algo)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(actual, expected), nil
}

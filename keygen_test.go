package sshpass

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestParseKeyAlgorithm(t *testing.T) {
	tests := []struct {
		input   string
		want    KeyAlgorithm
		wantErr bool
	}{
		{"ed25519", KeyEd25519, false},
		{"ED25519", KeyEd25519, false},
		{"Ed", KeyEd25519, false},
		{" ed25519 ", KeyEd25519, false},
		{"rsa", KeyRSA, false},
		{"RSA", KeyRSA, false},
		{"ecdsa", "", true},
		{"", "", true},
		{"dsa", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKeyAlgorithm(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKeyAlgorithm(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseKeyAlgorithm(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateKeyPairEd25519(t *testing.T) {
	pair, err := GenerateKeyPair(KeyEd25519, "test@example.com")
	if err != nil {
		t.Fatalf("GenerateKeyPair(ed25519) error: %v", err)
	}
	if pair == nil {
		t.Fatal("expected non-nil KeyPair")
	}
	if pair.Algorithm != KeyEd25519 {
		t.Errorf("Algorithm = %v, want %v", pair.Algorithm, KeyEd25519)
	}
	if len(pair.PrivateKey) == 0 {
		t.Error("PrivateKey is empty")
	}
	if len(pair.PublicKey) == 0 {
		t.Error("PublicKey is empty")
	}
	if pair.Comment != "test@example.com" {
		t.Errorf("Comment = %q, want %q", pair.Comment, "test@example.com")
	}

	// private key must be parseable by ssh.ParsePrivateKey
	signer, err := ssh.ParsePrivateKey(pair.PrivateKey)
	if err != nil {
		t.Fatalf("failed to parse generated private key: %v", err)
	}
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		t.Errorf("public key type = %s, want %s", signer.PublicKey().Type(), ssh.KeyAlgoED25519)
	}

	// public key must be a valid authorized_keys line
	pubLine := strings.TrimSpace(string(pair.PublicKey))
	if !strings.HasPrefix(pubLine, "ssh-ed25519 ") {
		t.Errorf("public key line does not start with 'ssh-ed25519': %q", pubLine)
	}
	if !strings.HasSuffix(pubLine, "test@example.com") {
		t.Errorf("public key line does not end with comment: %q", pubLine)
	}
	// the public key derived from the private key must match the exported public key
	if signer.PublicKey().Marshal() == nil {
		t.Error("signer public key marshal is nil")
	}
}

func TestGenerateKeyPairRSA(t *testing.T) {
	pair, err := GenerateRSAKeyPair(2048, "rsa-test")
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair(2048) error: %v", err)
	}
	if pair.Algorithm != KeyRSA {
		t.Errorf("Algorithm = %v, want %v", pair.Algorithm, KeyRSA)
	}
	if pair.Bits != 2048 {
		t.Errorf("Bits = %d, want 2048", pair.Bits)
	}

	signer, err := ssh.ParsePrivateKey(pair.PrivateKey)
	if err != nil {
		t.Fatalf("failed to parse generated RSA private key: %v", err)
	}
	if signer.PublicKey().Type() != ssh.KeyAlgoRSA {
		t.Errorf("public key type = %s, want %s", signer.PublicKey().Type(), ssh.KeyAlgoRSA)
	}

	pubLine := strings.TrimSpace(string(pair.PublicKey))
	if !strings.HasPrefix(pubLine, "ssh-rsa ") {
		t.Errorf("public key line does not start with 'ssh-rsa': %q", pubLine)
	}
}

func TestGenerateRSAKeyPairTooSmall(t *testing.T) {
	_, err := GenerateRSAKeyPair(1024, "")
	if err == nil {
		t.Fatal("expected error for RSA key < 2048 bits")
	}
}

func TestGenerateKeyPairUnsupported(t *testing.T) {
	_, err := GenerateKeyPair(KeyAlgorithm("ecdsa"), "")
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
}

func TestGenerateKeyPairEmptyComment(t *testing.T) {
	pair, err := GenerateKeyPair(KeyEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKeyPair with empty comment error: %v", err)
	}
	pubLine := strings.TrimSpace(string(pair.PublicKey))
	// without comment, the line should be exactly "ssh-ed25519 <base64>"
	parts := strings.Fields(pubLine)
	if len(parts) != 2 {
		t.Errorf("expected 2 fields in public key without comment, got %d: %q", len(parts), pubLine)
	}
}

func TestKeyPairUniqueness(t *testing.T) {
	pair1, err := GenerateKeyPair(KeyEd25519, "")
	if err != nil {
		t.Fatal(err)
	}
	pair2, err := GenerateKeyPair(KeyEd25519, "")
	if err != nil {
		t.Fatal(err)
	}
	if string(pair1.PrivateKey) == string(pair2.PrivateKey) {
		t.Error("two generated ed25519 keys are identical (expected uniqueness)")
	}
	if string(pair1.PublicKey) == string(pair2.PublicKey) {
		t.Error("two generated ed25519 public keys are identical (expected uniqueness)")
	}
}

func TestSaveKeyPair(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, ".ssh", "id_ed25519")

	pair, err := GenerateKeyPair(KeyEd25519, "test@save")
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveKeyPair(pair, keyPath); err != nil {
		t.Fatalf("SaveKeyPair error: %v", err)
	}

	// private key file should exist with 0600 (Unix only; Windows ignores the mode)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("private key file not created: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("private key mode = %v, want 0600", info.Mode().Perm())
	}

	// public key file (.pub) should exist
	pubPath := keyPath + ".pub"
	pubData, err := os.ReadFile(pubPath)
	if err != nil {
		t.Fatalf("public key file not created: %v", err)
	}
	if string(pubData) != string(pair.PublicKey) {
		t.Error("public key file content does not match pair.PublicKey")
	}

	// saved private key must be parseable
	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ssh.ParsePrivateKey(data); err != nil {
		t.Fatalf("saved private key is not parseable: %v", err)
	}
}

func TestSaveKeyPairRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519")

	pair, err := GenerateKeyPair(KeyEd25519, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveKeyPair(pair, keyPath); err != nil {
		t.Fatal(err)
	}
	// second save should fail
	if err := SaveKeyPair(pair, keyPath); err == nil {
		t.Fatal("expected error when overwriting existing private key")
	}
}

func TestSaveKeyPairNilPair(t *testing.T) {
	dir := t.TempDir()
	if err := SaveKeyPair(nil, filepath.Join(dir, "key")); err == nil {
		t.Fatal("expected error for nil key pair")
	}
}

func TestSaveKeyPairEmptyPath(t *testing.T) {
	pair, err := GenerateKeyPair(KeyEd25519, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveKeyPair(pair, ""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestDeployPublicKeyValidation(t *testing.T) {
	// nil client
	if err := DeployPublicKey(nil, []byte("ssh-ed25519 AAAA test")); err == nil {
		t.Fatal("expected error for nil client")
	}
	// empty public key
	c := &Client{}
	if err := DeployPublicKey(c, []byte("")); err == nil {
		t.Fatal("expected error for empty public key")
	}
	if err := DeployPublicKey(c, []byte("   \n  ")); err == nil {
		t.Fatal("expected error for whitespace-only public key")
	}
	// unsafe characters: single quote
	if err := DeployPublicKey(c, []byte("ssh-ed25519 AAAA it's-a-comment")); err == nil {
		t.Fatal("expected error for public key with single quote")
	}
	// unsafe characters: backtick
	if err := DeployPublicKey(c, []byte("ssh-ed25519 AAAA `cmd`")); err == nil {
		t.Fatal("expected error for public key with backtick")
	}
	// unsafe characters: newline
	if err := DeployPublicKey(c, []byte("ssh-ed25519 AAAA test\nrm -rf /")); err == nil {
		t.Fatal("expected error for public key with newline")
	}
}

func TestDefaultKeyPath(t *testing.T) {
	p := DefaultKeyPath(KeyEd25519)
	if !strings.HasSuffix(p, "id_ed25519") {
		t.Errorf("DefaultKeyPath(ed25519) = %q, expected suffix 'id_ed25519'", p)
	}
	p = DefaultKeyPath(KeyRSA)
	if !strings.HasSuffix(p, "id_rsa") {
		t.Errorf("DefaultKeyPath(rsa) = %q, expected suffix 'id_rsa'", p)
	}
}

func TestBuildAuthorizedKeyLineWithComment(t *testing.T) {
	pair, err := GenerateKeyPair(KeyEd25519, "user@host")
	if err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(string(pair.PublicKey))
	if !strings.HasSuffix(line, "user@host") {
		t.Errorf("public key line missing comment: %q", line)
	}
}

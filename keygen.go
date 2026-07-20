package sshpass

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// KeyAlgorithm represents the type of SSH key to generate.
type KeyAlgorithm string

const (
	// KeyEd25519 specifies an Ed25519 key pair (recommended: fast, small, secure).
	KeyEd25519 KeyAlgorithm = "ed25519"
	// KeyRSA specifies an RSA key pair (widely compatible, larger keys).
	KeyRSA KeyAlgorithm = "rsa"
)

// DefaultRSABits is the default bit size for generated RSA keys.
const DefaultRSABits = 4096

// MinRSABits is the minimum acceptable RSA key size.
const MinRSABits = 2048

// KeyPair holds a generated SSH key pair ready for writing to disk or deploying.
type KeyPair struct {
	// Algorithm is the key type (ed25519 or rsa).
	Algorithm KeyAlgorithm
	// PrivateKey is the PEM-encoded OpenSSH-format private key.
	PrivateKey []byte
	// PublicKey is the OpenSSH authorized_keys line (with trailing newline).
	PublicKey []byte
	// Comment is the comment embedded in the key (may be empty).
	Comment string
	// Bits is the RSA key size (0 for ed25519).
	Bits int
}

// ParseKeyAlgorithm parses a key algorithm name (case-insensitive).
// Supported values: "ed25519", "rsa". Returns an error for unknown algorithms.
func ParseKeyAlgorithm(s string) (KeyAlgorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ed25519", "ed":
		return KeyEd25519, nil
	case "rsa":
		return KeyRSA, nil
	default:
		return "", fmt.Errorf("unsupported key algorithm: %q (supported: ed25519, rsa)", s)
	}
}

// GenerateKeyPair generates a new SSH key pair using the given algorithm.
// algo must be KeyEd25519 or KeyRSA. comment is appended to the public key
// (typically "user@host"); it may be empty.
//
// The returned KeyPair.PrivateKey is PEM-encoded in the modern OpenSSH format.
// The returned KeyPair.PublicKey is a single authorized_keys line ending with
// a newline.
func GenerateKeyPair(algo KeyAlgorithm, comment string) (*KeyPair, error) {
	switch algo {
	case KeyEd25519:
		return generateEd25519Key(comment)
	case KeyRSA:
		return generateRSAKey(DefaultRSABits, comment)
	default:
		return nil, fmt.Errorf("unsupported key algorithm: %s (supported: ed25519, rsa)", algo)
	}
}

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
// bits must be at least MinRSABits (2048). comment is appended to the public key.
func GenerateRSAKeyPair(bits int, comment string) (*KeyPair, error) {
	if bits < MinRSABits {
		return nil, fmt.Errorf("RSA key size must be at least %d bits, got %d", MinRSABits, bits)
	}
	return generateRSAKey(bits, comment)
}

func generateEd25519Key(comment string) (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ed25519 private key: %w", err)
	}
	pemKey := pem.EncodeToMemory(pemBlock)

	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to create ed25519 public key: %w", err)
	}
	pubLine := buildAuthorizedKeyLine(pubKey, comment)

	return &KeyPair{
		Algorithm:  KeyEd25519,
		PrivateKey: pemKey,
		PublicKey:  []byte(pubLine + "\n"),
		Comment:    comment,
	}, nil
}

func generateRSAKey(bits int, comment string) (*KeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RSA private key: %w", err)
	}
	pemKey := pem.EncodeToMemory(pemBlock)

	pubKey, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA public key: %w", err)
	}
	pubLine := buildAuthorizedKeyLine(pubKey, comment)

	return &KeyPair{
		Algorithm:  KeyRSA,
		PrivateKey: pemKey,
		PublicKey:  []byte(pubLine + "\n"),
		Comment:    comment,
		Bits:       bits,
	}, nil
}

// buildAuthorizedKeyLine builds a single authorized_keys line from an ssh.PublicKey
// and an optional comment, without the trailing newline.
func buildAuthorizedKeyLine(pubKey ssh.PublicKey, comment string) string {
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))
	if comment != "" {
		line += " " + comment
	}
	return line
}

// SaveKeyPair writes the key pair to disk. privateKeyPath is the path for the
// private key (e.g. ~/.ssh/id_ed25519); the public key is written alongside it
// with a ".pub" suffix. The private key is created with 0600 permissions and
// the public key with 0644.
//
// If the private key file already exists, SaveKeyPair returns an error to avoid
// overwriting existing keys. Callers that wish to overwrite should delete the
// file first.
func SaveKeyPair(pair *KeyPair, privateKeyPath string) error {
	if pair == nil {
		return fmt.Errorf("key pair is nil")
	}
	if privateKeyPath == "" {
		return fmt.Errorf("private key path is empty")
	}

	// refuse to overwrite an existing private key
	if _, err := os.Stat(privateKeyPath); err == nil {
		return fmt.Errorf("private key file already exists: %s (remove it first to regenerate)", privateKeyPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check private key path: %w", err)
	}

	// create parent directory
	dir := filepath.Dir(privateKeyPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// write private key (0600)
	if err := os.WriteFile(privateKeyPath, pair.PrivateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// write public key (.pub, 0644)
	pubPath := privateKeyPath + ".pub"
	if err := os.WriteFile(pubPath, pair.PublicKey, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// DefaultKeyPath returns the default private key path for the given algorithm,
// based on the user's home directory (~/.ssh/id_ed25519 or ~/.ssh/id_rsa).
func DefaultKeyPath(algo KeyAlgorithm) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	var name string
	switch algo {
	case KeyRSA:
		name = "id_rsa"
	default:
		name = "id_ed25519"
	}
	return filepath.Join(homeDir, ".ssh", name)
}

// DeployPublicKey installs a public key on a remote host so that the
// corresponding private key can be used for password-less login.
//
// It connects through client (which must already be authenticated) and:
//  1. Creates ~/.ssh if it does not exist (mode 0700).
//  2. Appends the public key to ~/.ssh/authorized_keys if it is not already
//     present (idempotent — safe to call multiple times).
//  3. Sets ~/.ssh/authorized_keys to mode 0600.
//
// The public key must be a single authorized_keys line (with or without a
// trailing newline). Returns nil if the key was already present.
//
// This is a low-level SDK primitive. The CLI keygen command does NOT call this
// automatically — deployment is left to the user to avoid issues with complex
// server environments. Callers who need programmatic deployment can use this
// function directly.
func DeployPublicKey(client *Client, publicKey []byte) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	pubKeyLine := strings.TrimSpace(string(publicKey))
	if pubKeyLine == "" {
		return fmt.Errorf("public key is empty")
	}

	// OpenSSH public keys contain only base64, spaces, and optional comment
	// text — none of these include a single quote under normal circumstances.
	// We reject single quotes defensively to prevent shell-injection via the
	// comment field.
	if strings.ContainsAny(pubKeyLine, "'`\n\r") {
		return fmt.Errorf("public key contains unsafe characters (single quote, backtick, or newline)")
	}

	// Install the key using a remote shell script. grep -qxF performs an exact
	// full-line match so the key is only appended once.
	script := fmt.Sprintf(`umask 077
mkdir -p "$HOME/.ssh"
touch "$HOME/.ssh/authorized_keys"
grep -qxF '%s' "$HOME/.ssh/authorized_keys" 2>/dev/null || echo '%s' >> "$HOME/.ssh/authorized_keys"
chmod 700 "$HOME/.ssh"
chmod 600 "$HOME/.ssh/authorized_keys"
`, pubKeyLine, pubKeyLine)

	return client.Exec(script)
}

package sshpass

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFileAllAlgorithms(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(f, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		algo string
		want string
	}{
		{"md5", "6f5902ac237024bdd0c176cb93063dc4"},
		{"sha1", "22596363b3de40b06f981fb85d82312e8c0ed511"},
		{"sha256", "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"},
		{"sha512", "db3974a97f2407b7cae1ae637c0030687a11913274d578492558e39c16c017de84eacdc8c62fe34ee4e12b4b1428817f09b6a2760c3f8a664ceae94d2434a593"},
	}

	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			got, err := HashFile(f, tt.algo)
			if err != nil {
				t.Fatalf("HashFile(%q, %q) error: %v", f, tt.algo, err)
			}
			if got != tt.want {
				t.Errorf("HashFile(%q, %q) = %q, want %q", f, tt.algo, got, tt.want)
			}
		})
	}
}

func TestHashFileUnknownAlgorithm(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(f, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := HashFile(f, "blake3")
	if err == nil {
		t.Fatal("expected error for unknown algorithm")
	}
}

func TestHashFileNotFound(t *testing.T) {
	_, err := HashFile("/nonexistent/path/file.txt", "sha256")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestHashFileEmptyFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.bin")
	if err := os.WriteFile(f, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		algo string
		want string
	}{
		{"md5", "d41d8cd98f00b204e9800998ecf8427e"},
		{"sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		t.Run(tt.algo, func(t *testing.T) {
			got, err := HashFile(f, tt.algo)
			if err != nil {
				t.Fatalf("HashFile(%q, %q) error: %v", f, tt.algo, err)
			}
			if got != tt.want {
				t.Errorf("HashFile(%q, %q) = %q, want %q", f, tt.algo, got, tt.want)
			}
		})
	}
}

func TestVerifyFileMatch(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(f, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := VerifyFile(f, "sha256", "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447")
	if err != nil {
		t.Fatalf("VerifyFile error: %v", err)
	}
	if !ok {
		t.Fatal("expected match")
	}
}

func TestVerifyFileMismatch(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(f, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := VerifyFile(f, "sha256", "deadbeef")
	if err != nil {
		t.Fatalf("VerifyFile error: %v", err)
	}
	if ok {
		t.Fatal("expected mismatch")
	}
}

func TestVerifyFileCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(f, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := VerifyFile(f, "sha256", "A948904F2F0F479B8F8197694B30184B0D2ED1C1CD2A1EC0FB85D299A192A447")
	if err != nil {
		t.Fatalf("VerifyFile error: %v", err)
	}
	if !ok {
		t.Fatal("expected case-insensitive match")
	}
}

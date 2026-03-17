package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLocalFiles(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(dir, "configuration.nix"), []byte("config content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hardware.nix"), []byte("hardware content"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := readLocalFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files["configuration.nix"] != "config content" {
		t.Errorf("unexpected content for configuration.nix: %q", files["configuration.nix"])
	}
	if files["hardware.nix"] != "hardware content" {
		t.Errorf("unexpected content for hardware.nix: %q", files["hardware.nix"])
	}
}

func TestReadLocalFiles_Subdirectories(t *testing.T) {
	dir := t.TempDir()

	subDir := filepath.Join(dir, "modules")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "networking.nix"), []byte("networking"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := readLocalFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join("modules", "networking.nix")
	if _, ok := files[expected]; !ok {
		t.Errorf("expected file %q, got keys: %v", expected, keys(files))
	}
}

func TestReadLocalFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	files, err := readLocalFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestIsPathSafe(t *testing.T) {
	safe := []string{
		"configuration.nix",
		"hardware.nix",
		"flake.nix",
		"flake.lock",
		"modules/networking.nix",
		"hosts/web/default.nix",
	}
	for _, p := range safe {
		if !isPathSafe(p) {
			t.Errorf("expected %q to be safe", p)
		}
	}

	unsafe := []string{
		"",
		"../etc/passwd",
		"; rm -rf /",
		"$(whoami)",
		"`whoami`",
		"file name with spaces.nix",
		"/absolute/path",
		".hidden",
	}
	for _, p := range unsafe {
		if isPathSafe(p) {
			t.Errorf("expected %q to be unsafe", p)
		}
	}
}

func keys(m map[string]string) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}

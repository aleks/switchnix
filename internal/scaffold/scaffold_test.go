package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_CreatesStructure(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hostsPath := filepath.Join(dir, "hosts.yml")
	if _, err := os.Stat(hostsPath); os.IsNotExist(err) {
		t.Error("hosts.yml was not created")
	}

	configsPath := filepath.Join(dir, "configurations")
	info, err := os.Stat(configsPath)
	if os.IsNotExist(err) {
		t.Error("configurations/ directory was not created")
	} else if !info.IsDir() {
		t.Error("configurations is not a directory")
	}

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if len(content) == 0 {
		t.Error("hosts.yml is empty")
	}
}

func TestInit_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	hostsPath := filepath.Join(dir, "hosts.yml")
	if err := os.WriteFile(hostsPath, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Init(dir)
	if err == nil {
		t.Fatal("expected error when hosts.yml already exists")
	}

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Error("hosts.yml was overwritten")
	}
}

func TestInit_ExistingConfigurationsDir(t *testing.T) {
	dir := t.TempDir()
	configsPath := filepath.Join(dir, "configurations")
	if err := os.Mkdir(configsPath, 0755); err != nil {
		t.Fatal(err)
	}

	err := Init(dir)
	if err == nil {
		t.Fatal("expected error when configurations/ already exists")
	}
}

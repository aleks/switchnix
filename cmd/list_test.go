package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunList_WithHosts(t *testing.T) {
	dir := t.TempDir()
	hostsFile := filepath.Join(dir, "hosts.yml")
	content := []byte(`hosts:
  - name: web
    hostname: 10.0.0.1
    username: root
  - name: db
    hostname: 10.0.0.2
    username: admin
    port: 2222
`)
	if err := os.WriteFile(hostsFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Override configPath for the test
	oldConfigPath := configPath
	configPath = hostsFile
	defer func() { configPath = oldConfigPath }()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(nil, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("web")) {
		t.Error("expected output to contain 'web'")
	}
	if !bytes.Contains([]byte(output), []byte("db")) {
		t.Error("expected output to contain 'db'")
	}
	if !bytes.Contains([]byte(output), []byte("10.0.0.1")) {
		t.Error("expected output to contain '10.0.0.1'")
	}
	if !bytes.Contains([]byte(output), []byte("2222")) {
		t.Error("expected output to contain '2222'")
	}
}

func TestRunList_EmptyHosts(t *testing.T) {
	dir := t.TempDir()
	hostsFile := filepath.Join(dir, "hosts.yml")
	if err := os.WriteFile(hostsFile, []byte("hosts: []"), 0644); err != nil {
		t.Fatal(err)
	}

	oldConfigPath := configPath
	configPath = hostsFile
	defer func() { configPath = oldConfigPath }()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(nil, nil)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("No hosts configured")) {
		t.Error("expected 'No hosts configured' message")
	}
}

func TestRunList_MissingConfig(t *testing.T) {
	oldConfigPath := configPath
	configPath = "/nonexistent/hosts.yml"
	defer func() { configPath = oldConfigPath }()

	err := runList(nil, nil)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

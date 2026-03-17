package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.yml")
	content := []byte(`hosts:
  - name: webserver
    hostname: 192.168.1.10
    username: root
    port: 22
  - name: database
    hostname: db.example.com
    username: admin
    port: 2222
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(cfg.Hosts))
	}

	h := cfg.Hosts[0]
	if h.Name != "webserver" {
		t.Errorf("expected name 'webserver', got %q", h.Name)
	}
	if h.Hostname != "192.168.1.10" {
		t.Errorf("expected hostname '192.168.1.10', got %q", h.Hostname)
	}
	if h.Username != "root" {
		t.Errorf("expected username 'root', got %q", h.Username)
	}
	if h.Port != 22 {
		t.Errorf("expected port 22, got %d", h.Port)
	}

	h2 := cfg.Hosts[1]
	if h2.Port != 2222 {
		t.Errorf("expected port 2222, got %d", h2.Port)
	}
}

func TestLoadConfig_DefaultPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.yml")
	content := []byte(`hosts:
  - name: server
    hostname: 10.0.0.1
    username: nixos
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Hosts[0].Port != 22 {
		t.Errorf("expected default port 22, got %d", cfg.Hosts[0].Port)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/hosts.yml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.yml")
	if err := os.WriteFile(path, []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfig_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "empty name",
			content: `hosts:
  - name: ""
    hostname: 10.0.0.1
    username: root`,
		},
		{
			name: "missing hostname",
			content: `hosts:
  - name: server
    hostname: ""
    username: root`,
		},
		{
			name: "missing username",
			content: `hosts:
  - name: server
    hostname: 10.0.0.1
    username: ""`,
		},
		{
			name: "invalid name characters",
			content: `hosts:
  - name: "server;rm -rf /"
    hostname: 10.0.0.1
    username: root`,
		},
		{
			name: "invalid hostname characters",
			content: `hosts:
  - name: server
    hostname: "10.0.0.1; cat /etc/passwd"
    username: root`,
		},
		{
			name: "invalid username characters",
			content: `hosts:
  - name: server
    hostname: 10.0.0.1
    username: "root; whoami"`,
		},
		{
			name: "duplicate names",
			content: `hosts:
  - name: server
    hostname: 10.0.0.1
    username: root
  - name: server
    hostname: 10.0.0.2
    username: root`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "hosts.yml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := LoadConfig(path)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestFindHost(t *testing.T) {
	cfg := &Config{
		Hosts: []Host{
			{Name: "web", Hostname: "10.0.0.1", Username: "root", Port: 22},
			{Name: "db", Hostname: "10.0.0.2", Username: "admin", Port: 22},
		},
	}

	h, err := cfg.FindHost("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Hostname != "10.0.0.1" {
		t.Errorf("expected hostname '10.0.0.1', got %q", h.Hostname)
	}

	h, err = cfg.FindHost("db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Hostname != "10.0.0.2" {
		t.Errorf("expected hostname '10.0.0.2', got %q", h.Hostname)
	}

	_, err = cfg.FindHost("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent host")
	}
}

func TestValidName(t *testing.T) {
	valid := []string{"server", "web-01", "my_host", "Server123"}
	for _, name := range valid {
		if !isValidName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}

	invalid := []string{"", "server;cmd", "host name", "a/b", "$(whoami)", "host\nname"}
	for _, name := range invalid {
		if isValidName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}

func TestValidHostname(t *testing.T) {
	valid := []string{"192.168.1.1", "example.com", "my-server.local", "10.0.0.1", "server01"}
	for _, h := range valid {
		if !isValidHostname(h) {
			t.Errorf("expected %q to be valid hostname", h)
		}
	}

	invalid := []string{"", "host;cmd", "host name", "$(whoami)", "host\n"}
	for _, h := range invalid {
		if isValidHostname(h) {
			t.Errorf("expected %q to be invalid hostname", h)
		}
	}
}

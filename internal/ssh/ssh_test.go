package ssh

import (
	"strings"
	"testing"

	"github.com/aleks/switchnix/internal/config"
)

func TestSSHArgs_DefaultPort(t *testing.T) {
	h := &config.Host{
		Name:     "web",
		Hostname: "10.0.0.1",
		Username: "root",
		Port:     22,
	}

	args := SSHArgs(h)
	expected := []string{"-o", "BatchMode=yes", "-p", "22", "root@10.0.0.1"}
	assertSliceEqual(t, expected, args)
}

func TestSSHArgs_CustomPort(t *testing.T) {
	h := &config.Host{
		Name:     "db",
		Hostname: "db.example.com",
		Username: "admin",
		Port:     2222,
	}

	args := SSHArgs(h)
	expected := []string{"-o", "BatchMode=yes", "-p", "2222", "admin@db.example.com"}
	assertSliceEqual(t, expected, args)
}

func TestSSHArgs_WithSSHOptions(t *testing.T) {
	h := &config.Host{
		Name:       "web",
		Hostname:   "10.0.0.1",
		Username:   "root",
		Port:       22,
		SSHOptions: []string{"-o", "ConnectTimeout=5"},
	}

	args := SSHArgs(h)
	expected := []string{"-o", "BatchMode=yes", "-p", "22", "-o", "ConnectTimeout=5", "root@10.0.0.1"}
	assertSliceEqual(t, expected, args)
}

func TestInteractiveSSHArgs(t *testing.T) {
	h := &config.Host{
		Name:     "web",
		Hostname: "10.0.0.1",
		Username: "root",
		Port:     22,
	}

	args := InteractiveSSHArgs(h)
	expected := []string{"-t", "-p", "22", "root@10.0.0.1"}
	assertSliceEqual(t, expected, args)
}

func TestRsyncArgs_Pull(t *testing.T) {
	h := &config.Host{
		Name:     "web",
		Hostname: "10.0.0.1",
		Username: "root",
		Port:     22,
	}

	args := RsyncPullArgs(h, "/tmp/switchnix-stage/", "/local/path/")
	assertContains(t, args, "-avz")
	assertNotContains(t, args, "--rsync-path=sudo rsync")
	assertContains(t, args, "root@10.0.0.1:/tmp/switchnix-stage/")
	assertContains(t, args, "/local/path/")

	// Should contain ssh command with BatchMode and port
	found := false
	for i, a := range args {
		if a == "-e" && i+1 < len(args) {
			if args[i+1] == "ssh -o BatchMode=yes -p 22" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected rsync args to contain -e with ssh command, got %v", args)
	}
}

func TestRsyncArgs_PullWithSSHOptions(t *testing.T) {
	h := &config.Host{
		Name:       "web",
		Hostname:   "10.0.0.1",
		Username:   "root",
		Port:       22,
		SSHOptions: []string{"-o", "ConnectTimeout=5"},
	}

	args := RsyncPullArgs(h, "/tmp/switchnix-stage/", "/local/path/")
	found := false
	for i, a := range args {
		if a == "-e" && i+1 < len(args) {
			v := args[i+1]
			if strings.Contains(v, "BatchMode=yes") &&
				strings.Contains(v, "ConnectTimeout=5") &&
				strings.Contains(v, "-p 22") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected rsync args to contain ssh command with SSHOptions, got %v", args)
	}
}

func TestRsyncArgs_Push(t *testing.T) {
	h := &config.Host{
		Name:     "web",
		Hostname: "10.0.0.1",
		Username: "root",
		Port:     22,
	}

	args := RsyncPushArgs(h, "/local/path/", "/tmp/switchnix-stage/")
	assertContains(t, args, "-avz")
	assertContains(t, args, "--delete")
	assertNotContains(t, args, "--rsync-path=sudo rsync")
	assertContains(t, args, "/local/path/")
	assertContains(t, args, "root@10.0.0.1:/tmp/switchnix-stage/")
}

func assertSliceEqual(t *testing.T, expected, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("at index %d: expected %q, got %q\nfull: expected %v, got %v", i, expected[i], actual[i], expected, actual)
		}
	}
}

func assertContains(t *testing.T, slice []string, value string) {
	t.Helper()
	for _, s := range slice {
		if s == value {
			return
		}
	}
	t.Errorf("expected slice to contain %q, got %v", value, slice)
}

func assertNotContains(t *testing.T, slice []string, value string) {
	t.Helper()
	for _, s := range slice {
		if s == value {
			t.Errorf("expected slice NOT to contain %q, got %v", value, slice)
			return
		}
	}
}

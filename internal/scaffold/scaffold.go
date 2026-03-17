package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

const hostsTemplate = `# switchnix hosts configuration
# Define your NixOS remote hosts here.
#
# hosts:
#   - name: webserver          # Identifier used in CLI commands
#     hostname: 192.168.1.10   # IP address or hostname
#     username: root           # SSH username
#     port: 22                 # SSH port (optional, default: 22)
#     ssh_options: []          # Extra SSH flags (optional)

hosts: []
`

// Init creates the switchnix project structure in the given directory.
func Init(dir string) error {
	hostsPath := filepath.Join(dir, "hosts.yml")
	configsPath := filepath.Join(dir, "configurations")

	if _, err := os.Stat(hostsPath); err == nil {
		return fmt.Errorf("hosts.yml already exists in %s", dir)
	}
	if _, err := os.Stat(configsPath); err == nil {
		return fmt.Errorf("configurations/ already exists in %s", dir)
	}

	if err := os.WriteFile(hostsPath, []byte(hostsTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create hosts.yml: %w", err)
	}

	if err := os.MkdirAll(configsPath, 0755); err != nil {
		return fmt.Errorf("failed to create configurations/: %w", err)
	}

	return nil
}

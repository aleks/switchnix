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
// Uses O_CREATE|O_EXCL to atomically check-and-create, avoiding TOCTOU races.
func Init(dir string) error {
	hostsPath := filepath.Join(dir, "hosts.yml")
	configsPath := filepath.Join(dir, "configurations")

	// Atomically create hosts.yml — fails if it already exists
	f, err := os.OpenFile(hostsPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("hosts.yml already exists in %s", dir)
		}
		return fmt.Errorf("failed to create hosts.yml: %w", err)
	}
	if _, err := f.WriteString(hostsTemplate); err != nil {
		f.Close()
		return fmt.Errorf("failed to write hosts.yml: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close hosts.yml: %w", err)
	}

	// Mkdir also fails atomically if the directory already exists
	if err := os.Mkdir(configsPath, 0755); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("configurations/ already exists in %s", dir)
		}
		return fmt.Errorf("failed to create configurations/: %w", err)
	}

	return nil
}

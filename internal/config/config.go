package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Host struct {
	Name       string   `yaml:"name"`
	Hostname   string   `yaml:"hostname"`
	Username   string   `yaml:"username"`
	Port       int      `yaml:"port,omitempty"`
	SSHOptions []string `yaml:"ssh_options,omitempty"`
}

type Config struct {
	Hosts []Host `yaml:"hosts"`
}

var (
	nameRegexp     = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	hostnameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
	usernameRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
	// sshOptionRegexp allows only flags starting with - and safe values.
	sshOptionRegexp = regexp.MustCompile(`^-[a-zA-Z0-9]$|^[a-zA-Z0-9][a-zA-Z0-9._=:/-]*$`)
)

func isValidName(s string) bool {
	return nameRegexp.MatchString(s)
}

func isValidHostname(s string) bool {
	// Check for IPv6 address (with or without brackets)
	stripped := strings.TrimPrefix(strings.TrimSuffix(s, "]"), "[")
	if net.ParseIP(stripped) != nil {
		return true
	}
	// Fall back to hostname regex for DNS names and IPv4
	return hostnameRegexp.MatchString(s)
}

func isValidUsername(s string) bool {
	return usernameRegexp.MatchString(s)
}

func isValidSSHOption(s string) bool {
	return sshOptionRegexp.MatchString(s)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	seen := make(map[string]bool)
	for i := range cfg.Hosts {
		h := &cfg.Hosts[i]

		if h.Port == 0 {
			h.Port = 22
		}

		if err := validateHost(h); err != nil {
			return nil, fmt.Errorf("host %d: %w", i, err)
		}

		lower := strings.ToLower(h.Name)
		if seen[lower] {
			return nil, fmt.Errorf("duplicate host name: %q", h.Name)
		}
		seen[lower] = true
	}

	return &cfg, nil
}

func validateHost(h *Host) error {
	if !isValidName(h.Name) {
		return fmt.Errorf("invalid name %q: must be alphanumeric, hyphens, or underscores", h.Name)
	}
	if !isValidHostname(h.Hostname) {
		return fmt.Errorf("invalid hostname %q: must be a valid hostname, IPv4, or IPv6 address", h.Hostname)
	}
	if !isValidUsername(h.Username) {
		return fmt.Errorf("invalid username %q: must be alphanumeric, hyphens, underscores, or dots", h.Username)
	}
	if h.Port < 1 || h.Port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", h.Port)
	}
	for _, opt := range h.SSHOptions {
		if !isValidSSHOption(opt) {
			return fmt.Errorf("invalid ssh_option %q: contains disallowed characters", opt)
		}
	}
	return nil
}

func (c *Config) FindHost(name string) (*Host, error) {
	lower := strings.ToLower(name)
	for i := range c.Hosts {
		if strings.ToLower(c.Hosts[i].Name) == lower {
			return &c.Hosts[i], nil
		}
	}

	available := make([]string, len(c.Hosts))
	for i, h := range c.Hosts {
		available[i] = h.Name
	}
	return nil, fmt.Errorf("host %q not found. Available hosts: %s", name, strings.Join(available, ", "))
}

# Getting Started with switchnix

## Installation

```bash
go install github.com/aleks/switchnix@latest
```

Or build from source:

```bash
git clone https://github.com/aleks/switchnix.git
cd switchnix
go build -ldflags "-X github.com/aleks/switchnix/cmd.Version=1.0.0" -o switchnix .
```

## Version

```bash
switchnix --version
```

The version is set at build time via `-ldflags`. Development builds show `dev`.

## Quick start

```bash
# 1. Initialize a new project
mkdir my-nixos-configs && cd my-nixos-configs
switchnix init

# 2. Edit hosts.yml to add your servers
cat > hosts.yml <<EOF
hosts:
  - name: webserver
    hostname: 192.168.1.10
    username: root
EOF

# 3. Pull the current configuration from the server
switchnix pull webserver

# 4. Edit the configuration locally
vim configurations/webserver/configuration.nix

# 5. Push and apply atomically
switchnix switch webserver
```

## Prerequisites

- **Go 1.21+** for building from source
- **rsync** installed locally and on remote hosts
- **SSH key-based authentication** configured for your remote hosts (via ssh-agent, 1Password SSH agent, etc.)
- Remote user must have **sudo** access on NixOS hosts

## Global flags

| Flag | Default | Description |
|---|---|---|
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |
| `--version` | — | Print version and exit |
| `--help, -h` | — | Show help |

## Commands

| Command | Description |
|---|---|
| [`init`](init.md) | Initialize a new switchnix project |
| [`list`](list.md) | List configured hosts |
| [`pull`](pull.md) | Pull NixOS configuration from a remote host |
| [`push`](push.md) | Push local configuration to a remote host |
| [`switch`](switch.md) | Apply configuration on a remote host via nixos-rebuild |

## Security

switchnix is designed to avoid introducing security risks:

- **No credential storage**: Relies entirely on SSH agent/key-based authentication
- **No password capture**: Sudo prompts pass through directly via PTY
- **Host key verification**: Delegated to system SSH (`known_hosts`)
- **Input validation**: All user-provided values (hostnames, usernames, names, SSH options) are validated against strict allowlists
- **Path safety**: Remote file paths are validated to prevent command injection
- **BatchMode**: Non-interactive SSH commands use `BatchMode=yes` to prevent password fallback
- **File size limits**: Local files > 10 MB are rejected during push
- **Signal handling**: Ctrl+C cleanly cancels all operations

# switchnix switch

Apply the NixOS configuration on a remote host by running `nixos-rebuild`.

## Usage

```
switchnix switch <host>
switchnix switch <host> --action test
switchnix switch <host> --action boot
```

Where `<host>` is the name of a host defined in `hosts.yml`.

## What it does

1. Connects to the remote host via SSH with a PTY (pseudo-terminal)
2. Runs `sudo nixos-rebuild <action>` on the remote host
3. Streams all output (build progress, errors, warnings) directly to your terminal
4. Supports interactive `sudo` password prompts — you type your password directly

## Flags

| Flag | Default | Description |
|---|---|---|
| `--action` | `switch` | The nixos-rebuild action to run |
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |

### Available actions

| Action | Description |
|---|---|
| `switch` | Build and activate the configuration immediately, and make it the boot default |
| `test` | Build and activate the configuration immediately, but do not make it the boot default |
| `boot` | Build the configuration and make it the boot default, but do not activate until next reboot |

## How it works

The command runs `ssh -t` which allocates a real PTY on the remote host. This means:

- **sudo prompts** are forwarded to your terminal — type your password directly
- **Build output** streams in real time, just as if you were running the command locally
- **Ctrl+C** is properly forwarded to cancel the remote operation
- The SSH process's stdin/stdout/stderr are connected directly to your terminal

Unlike the `pull` and `push` commands, the `switch` command does **not** use `BatchMode=yes` because it needs interactive terminal support for sudo.

There is no timeout for the `switch` command since `nixos-rebuild` can take a long time (downloading packages, building derivations). Ctrl+C cleanly cancels the operation via signal handling.

## Typical workflow

```bash
# 1. Pull current configuration
switchnix pull webserver

# 2. Edit configuration locally
vim configurations/webserver/configuration.nix

# 3. Review and push changes
switchnix push webserver

# 4. Apply the configuration
switchnix switch webserver
```

## Examples

```bash
# Apply configuration immediately (default)
switchnix switch webserver

# Test configuration without making it the boot default
switchnix switch webserver --action test

# Set as boot default without activating now
switchnix switch webserver --action boot
```

## Prerequisites

- SSH access to the remote host (key-based or password auth)
- The remote user must have `sudo` access
- The NixOS configuration on the remote must be valid (push first if you've made changes)

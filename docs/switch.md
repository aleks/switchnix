# switchnix switch

Atomically push and apply the NixOS configuration on a remote host.

## Usage

```
switchnix switch <host>
switchnix switch <host> --action test
switchnix switch <host> --no-push
switchnix switch <host> --dry-run
switchnix switch <host> --nixos-args='--flake,/etc/nixos#myhost'
```

Where `<host>` is the name of a host defined in `hosts.yml`.

## What it does

By default, `switch` performs an **atomic push+rebuild**:

1. Reads local and remote configurations and shows a diff
2. Prompts for confirmation
3. Rsyncs local config to a staging directory (`/tmp/switchnix-stage/`) on the remote host
4. Runs `sudo nixos-rebuild <action> -I nixos-config=/tmp/switchnix-stage/configuration.nix` against the staged config
5. **If the rebuild succeeds**: commits the staged files to `/etc/nixos/` and cleans up
6. **If the rebuild fails**: cleans up the staging directory — `/etc/nixos/` is untouched

This means a broken configuration will never overwrite your working `/etc/nixos/`.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--action` | `switch` | The nixos-rebuild action to run |
| `--no-push` | `false` | Skip pushing; rebuild from the current remote `/etc/nixos/` |
| `--dry-run` | `false` | Show diff without pushing or switching |
| `--nixos-args` | `[]` | Additional flags to pass to `nixos-rebuild` |
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |

### Available actions

| Action | Description |
|---|---|
| `switch` | Build and activate the configuration immediately, and make it the boot default |
| `test` | Build and activate the configuration immediately, but do not make it the boot default |
| `boot` | Build the configuration and make it the boot default, but do not activate until next reboot |

## How it works

The default flow (without `--no-push`) combines `push` and `switch` into a single atomic operation. The configuration is first rsynced to a staging directory, and `nixos-rebuild` runs against that staging directory. Only after a successful rebuild are the files committed to `/etc/nixos/`. If the rebuild fails, the staging directory is cleaned up and `/etc/nixos/` remains unchanged.

With `--no-push`, the command simply runs `sudo nixos-rebuild <action>` against the existing `/etc/nixos/` — the same behavior as the standalone `switch` command before this feature was added.

The command runs `ssh -t` which allocates a real PTY on the remote host. This means:

- **sudo prompts** are forwarded to your terminal — type your password directly
- **Build output** streams in real time, just as if you were running the command locally
- **Ctrl+C** is properly forwarded to cancel the remote operation

There is no timeout for the `switch` command since `nixos-rebuild` can take a long time (downloading packages, building derivations). Ctrl+C cleanly cancels the operation via signal handling.

## Typical workflow

```bash
# 1. Pull current configuration
switchnix pull webserver

# 2. Edit configuration locally
vim configurations/webserver/configuration.nix

# 3. Push and apply atomically
switchnix switch webserver
```

## Examples

```bash
# Push and apply configuration (default)
switchnix switch webserver

# Preview changes without pushing or switching
switchnix switch webserver --dry-run

# Test configuration without making it the boot default
switchnix switch webserver --action test

# Just rebuild from the existing remote config (no push)
switchnix switch webserver --no-push

# Set as boot default without activating now
switchnix switch webserver --action boot

# Use a flake-based configuration
switchnix switch webserver --nixos-args='--flake,/etc/nixos#webserver'

# Pass multiple extra flags
switchnix switch webserver --nixos-args=--flake --nixos-args=/etc/nixos#webserver
```

## Prerequisites

- SSH access to the remote host (key-based or password auth)
- The remote user must have `sudo` access
- A local configuration directory must exist (run `switchnix pull <host>` first)

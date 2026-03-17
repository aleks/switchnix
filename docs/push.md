# switchnix push

Push local NixOS configuration to a remote host, with diff preview and confirmation.

## Usage

```
switchnix push <host>
switchnix push <host> --dry-run
```

Where `<host>` is the name of a host defined in `hosts.yml`.

## What it does

1. Reads the local configuration from `configurations/<host>/`
2. Fetches the current remote configuration from `/etc/nixos/` via SSH
3. Computes and displays a colored unified diff showing:
   - Files to be **added** (exist locally but not remotely)
   - Files to be **removed** (exist remotely but not locally)
   - Files to be **modified** (content differs)
4. Prompts for confirmation: `Apply these changes? [y/N]` (default is No)
5. On confirmation, pushes the local configuration to the remote using `rsync --delete`

## Flags

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Show the diff without pushing any changes |
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |

## Safety features

- **Diff preview**: Always shows exactly what will change before any modification
- **Explicit confirmation**: Defaults to No — you must type `y` or `yes` to proceed
- **Dry run mode**: Use `--dry-run` to inspect changes without risk
- **Path validation**: Remote file paths are validated against a strict allowlist to prevent command injection
- **Key-based auth only**: SSH uses `BatchMode=yes` for file fetching (no password fallback)

## How it works

The diff is computed by fetching each remote file individually via `ssh sudo cat`, then comparing with local file contents in memory. This avoids creating temporary files and works reliably with small NixOS configurations.

The actual push uses `rsync --delete` with `sudo rsync` on the remote, which means files removed locally will also be removed from the remote.

## Examples

```bash
# Preview changes without pushing
switchnix push webserver --dry-run

# Push configuration to webserver
switchnix push webserver
# Shows diff, then:
# Apply these changes to webserver (192.168.1.10)? [y/N]: y
```

## Prerequisites

- `configurations/<host>/` must exist (run `switchnix pull <host>` first)
- `rsync` must be installed locally and on the remote
- SSH key-based authentication must be configured
- The remote user must have `sudo` access

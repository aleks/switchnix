# switchnix pull

Pull NixOS configuration files from a remote host into the local project.

## Usage

```
switchnix pull <host>
```

Where `<host>` is the name of a host defined in `hosts.yml`.

## What it does

1. Connects to the remote host via SSH
2. Copies all files from `/etc/nixos/` on the remote into `configurations/<host>/` locally
3. Creates the `configurations/<host>/` directory if it doesn't exist

Typical files pulled include `configuration.nix`, `hardware-configuration.nix`, `flake.nix`, `flake.lock`, and any module files in subdirectories.

## How it works

Uses `rsync` over SSH to efficiently copy files. The remote rsync runs with `sudo` to read root-owned NixOS configuration files.

The SSH connection uses `BatchMode=yes`, meaning it relies entirely on SSH key/agent authentication and will fail immediately if key auth is not available (it will never prompt for a password).

## Flags

| Flag | Default | Description |
|---|---|---|
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |

## Timeouts and signals

The rsync operation has a 10-minute timeout. Ctrl+C cleanly cancels the operation.

## Prerequisites

- `rsync` must be installed locally and on the remote host (available by default on NixOS)
- SSH key-based authentication must be configured for the remote host
- The remote user must have `sudo` access (or be `root`)

## Examples

```bash
# Pull configuration from a host named "webserver"
switchnix pull webserver

# Use a custom config file
switchnix pull webserver --config /path/to/hosts.yml
```

## File structure after pull

```
configurations/
  webserver/
    configuration.nix
    hardware-configuration.nix
    flake.nix           # if the host uses flakes
    flake.lock          # if the host uses flakes
    modules/            # any subdirectories are preserved
      networking.nix
```

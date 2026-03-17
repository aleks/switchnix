# switchnix

A simple CLI tool for managing NixOS configurations on remote hosts from any machine — no Nix installation required locally.

## Why switchnix?

Most NixOS deployment tools (NixOps, Colmena, deploy-rs, morph) assume you're working within the Nix ecosystem: they evaluate Nix expressions locally, build derivations, and push closures to remote hosts. This is powerful, but it also means you need Nix installed on your local machine, often with flakes configured, and your entire configuration structured as a Nix project.

switchnix was built to solve a different, simpler problem: **managing NixOS configuration files on a small set of remote hosts from a machine that doesn't have Nix installed.** Maybe you're on macOS without Nix, editing from a work laptop, or you simply prefer to keep your `.nix` files in a plain Git repo and let the remote hosts handle the actual builds.

switchnix treats NixOS configuration as plain files. It pulls them, lets you edit them locally with whatever tools you like, shows you a diff of what changed, pushes them back, and triggers `nixos-rebuild` on the remote — all over SSH.

## How it compares

| | switchnix | NixOps | Colmena | deploy-rs | morph |
|---|---|---|---|---|---|
| Local Nix required | No | Yes | Yes | Yes | Yes |
| Flakes required | No | No | Optional | Yes | No |
| Builds happen | On remote | Local or remote | Local | Local | Local |
| State management | Stateless | Stateful | Stateless | Stateless | Stateless |
| Cloud provisioning | No | Yes | No | No | No |
| Parallel deploys | No | Yes | Yes | Yes | Yes |
| Scope | Small fleet | Any scale | Any scale | Any scale | Any scale |

**Use switchnix when:**
- You don't have (or want) Nix installed locally
- You manage a handful of NixOS machines, not dozens
- You want to edit `.nix` files with your normal editor and tools
- You want a clear diff before anything touches the remote
- You prefer simplicity over automation

**Use something else when:**
- You want to evaluate and build Nix expressions locally before deploying
- You need to manage a large fleet with parallel deployments
- You want to provision cloud infrastructure as part of your workflow

## Installation

```sh
go install github.com/aleks/switchnix@latest
```

Or build from source:

```sh
git clone https://github.com/aleks/switchnix.git
cd switchnix
go build -o switchnix .
```

## Quick start

```sh
# Initialize a new project
switchnix init

# Edit hosts.yml to add your NixOS hosts
vim hosts.yml

# Pull the current configuration from a remote host
switchnix pull myhost

# Edit the configuration locally
vim configurations/myhost/configuration.nix

# Preview what would change
switchnix switch myhost --dry-run

# Push and apply atomically (with diff and confirmation prompt)
switchnix switch myhost
```

## Commands

### `switchnix init`

Creates a `hosts.yml` config file and a `configurations/` directory in the current folder.

### `switchnix list`

Displays all configured hosts in a table.

### `switchnix pull <host>`

Pulls `/etc/nixos/` from the remote host into `configurations/<host>/` locally. Prompts for the sudo password interactively if needed.

### `switchnix push <host> [--dry-run]`

Compares local files against the remote `/etc/nixos/`, displays a colored unified diff, and asks for confirmation before pushing. Use `--dry-run` to preview without applying. Does **not** run `nixos-rebuild` — use `switch` for that.

### `switchnix switch <host> [--action switch|test|boot] [--no-push] [--dry-run] [--nixos-args=...]`

Atomically pushes local configuration and runs `nixos-rebuild` on the remote host. Files are staged in a temporary directory and `nixos-rebuild` runs against the staged config. Only if the rebuild succeeds are files committed to `/etc/nixos/`. If the rebuild fails, `/etc/nixos/` is left untouched.

Use `--no-push` to skip pushing and rebuild from the existing remote `/etc/nixos/`. Use `--dry-run` to preview the diff without pushing or switching. Use `--nixos-args` to pass additional flags to `nixos-rebuild` (e.g. `--nixos-args='--flake,/etc/nixos#myhost'`).

## Configuration

Hosts are defined in `hosts.yml`:

```yaml
hosts:
  - name: webserver
    hostname: 192.168.1.10
    username: deploy
    port: 22
    ssh_options:
      - "-o"
      - "ConnectTimeout=10"
```

| Field | Required | Default | Description |
|---|---|---|---|
| `name` | Yes | — | Unique host identifier |
| `hostname` | Yes | — | IP address (v4/v6) or DNS name |
| `username` | Yes | — | SSH username |
| `port` | No | `22` | SSH port |
| `ssh_options` | No | `[]` | Additional SSH flags |

## Prerequisites

- **Local:** `rsync` and `ssh` available in `$PATH`
- **Remote:** `rsync` installed, SSH key-based authentication, user with `sudo` access

## License

MIT

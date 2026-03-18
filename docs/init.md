# switchnix init

Initialize a new switchnix project in the current directory.

## Usage

```
switchnix init
```

## What it does

Creates the following structure in the current directory:

```
hosts.yml          - Host configuration file with commented examples
configurations/    - Directory for storing per-host NixOS configurations
```

The generated `hosts.yml` contains commented examples showing the expected format. The `configurations/` directory starts empty — use `switchnix pull` to populate it from remote servers.

## Safety

- Refuses to run if `hosts.yml` or `configurations/` already exist, preventing accidental overwrites.
- Uses atomic file creation (`O_EXCL`) to prevent race conditions if two `init` commands run simultaneously.

## hosts.yml format

After running `init`, edit `hosts.yml` to define your remote NixOS hosts:

```yaml
hosts:
  - name: webserver          # Identifier used in CLI commands
    hostname: 192.168.1.10   # IP address or hostname
    username: root           # SSH username
    port: 22                 # SSH port (optional, default: 22)
    ssh_options: []          # Extra SSH flags (optional)
    switch_args: ""          # Extra args for nixos-rebuild (optional)
```

### Field reference

| Field | Required | Default | Description |
|---|---|---|---|
| `name` | Yes | — | Unique identifier for the host. Used as CLI argument and directory name. Must be alphanumeric, hyphens, or underscores. |
| `hostname` | Yes | — | IP address (IPv4 or IPv6), or DNS hostname of the remote server. IPv6 addresses may be bare (`::1`) or bracketed (`[::1]`). |
| `username` | Yes | — | SSH username for connecting to the host. |
| `port` | No | `22` | SSH port. |
| `ssh_options` | No | `[]` | Additional flags passed to the `ssh` command (e.g., `["-o", "ConnectTimeout=5"]`). Validated against an allowlist to prevent injection. |
| `switch_args` | No | `""` | Additional arguments passed to `nixos-rebuild` when running `switchnix switch` (e.g., `"--flake /etc/nixos#webserver"`). Validated to prevent shell injection. |

## Next steps

1. Add your hosts to `hosts.yml`
2. Run `switchnix pull <host>` to fetch existing configurations

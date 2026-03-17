# switchnix list

Display all configured hosts.

## Usage

```
switchnix list
```

## What it does

Reads `hosts.yml` and prints a table of all configured hosts with their name, hostname, username, and port.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--config, -c` | `hosts.yml` | Path to hosts configuration file |

## Example output

```
NAME                 HOSTNAME                       USERNAME        PORT
webserver            192.168.1.10                   root            22
database             db.example.com                 admin           2222
```

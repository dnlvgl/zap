# zap

A TUI for killing processes by port number. Detects containers (Podman/Docker) and systemd services for proper shutdown.

## Install

```bash
go install github.com/dnl/zap/cmd/zap@latest
```

Or build from source:

```bash
go build -o zap ./cmd/zap/
```

## Usage

```bash
# Interactive TUI showing all listening ports
zap

# Target a specific port
zap :3000

# Port range
zap :8080-8090

# Specific interface
zap localhost:5432

# Force kill (SIGKILL)
zap :3000 --force

# Dry run (non-interactive, shows what would be killed)
zap :3000 --dry-run
```

## TUI

The default mode launches a fullscreen terminal interface.

Each process shows:
- PID and port/protocol
- Command line
- Memory usage, uptime, child process count
- Tags for container runtime (podman/docker) and systemd unit

### Key bindings

| Key | Action |
|-----|--------|
| `j` / `k` / arrows | Navigate |
| `enter` / `space` | Select process to kill |
| `y` / `enter` | Confirm kill |
| `n` / `esc` | Cancel |
| `r` | Refresh |
| `q` / `ctrl+c` | Quit |

## Kill strategies

zap automatically picks the best way to stop a process:

1. **Container** — `podman stop` / `docker stop` for containerized processes
2. **Systemd** — `systemctl stop` for systemd-managed services
3. **Signal** — `SIGTERM` (or `SIGKILL` with `--force`) for bare processes

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Use SIGKILL / container kill instead of graceful stop |
| `--dry-run` | `-n` | Show what would be killed (non-interactive) |
| `--version` | `-v` | Print version |
| `--help` | `-h` | Show help |

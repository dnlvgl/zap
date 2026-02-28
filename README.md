# zap

Too many dev servers running and ports colliding? Can't remember the `netstat` incantation to figure out what's hogging port 3000?

zap gives you a live TUI to find processes by port and kill them — with proper handling for containers (Podman/Docker) and systemd services. The list auto-refreshes every 2 seconds so you can watch a service come up, confirm a kill took effect, or spot new port conflicts without pressing a key.

![zap screenshot](screenshots/zap-screenshot.png)

## Install

```bash
go install github.com/dnlvgl/zap/cmd/zap@latest
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

## Releasing a new version

Pushing a `v*` tag triggers the GitHub Actions release workflow, which uses GoReleaser to build and publish binaries.

```bash
git tag v0.2.0
git push origin v0.2.0
```

package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnlvgl/zap/internal/container"
	"github.com/dnlvgl/zap/internal/kill"
	"github.com/dnlvgl/zap/internal/port"
	"github.com/dnlvgl/zap/internal/process"
	"github.com/dnlvgl/zap/internal/ui"
)

var version = "dev" // overridden at build time via -ldflags

type options struct {
	force   bool
	dryRun  bool
	version bool
	ports   []string
}

func parseArgs(args []string) options {
	var opts options
	for _, arg := range args {
		switch arg {
		case "--force", "-f":
			opts.force = true
		case "--dry-run", "-n":
			opts.dryRun = true
		case "--version", "-v":
			opts.version = true
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
				os.Exit(1)
			}
			opts.ports = append(opts.ports, arg)
		}
	}
	return opts
}

func printUsage() {
	fmt.Print(`Usage: zap [flags] [port...]

Kill processes by port number.

Arguments:
  port          Port to target (e.g. :3000, :8080-8090, localhost:5432)
                If omitted, lists all listening ports.

Flags:
  -f, --force     Use SIGKILL instead of SIGTERM
  -n, --dry-run   Show what would be killed without doing it
  -v, --version   Print version and exit
  -h, --help      Show this help
`)
}

func main() {
	opts := parseArgs(os.Args[1:])

	if opts.version {
		fmt.Println("zap " + version)
		os.Exit(0)
	}

	// Dry-run mode: non-interactive text output
	if opts.dryRun {
		runDryRun(opts)
		return
	}

	// Interactive TUI mode
	var query *port.Query
	if len(opts.ports) > 0 {
		// For now, use the first port argument for TUI
		q, err := port.Parse(opts.ports[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		query = &q
	}

	model := ui.New(query, opts.force)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runDryRun(opts options) {
	queries := opts.ports
	if len(queries) == 0 {
		// Dry-run with no ports: show all
		queries = []string{"1-65535"}
	}

	hasError := false
	for _, arg := range queries {
		q, err := port.Parse(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		listeners, err := port.Detect(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error detecting processes on %s: %v\n", arg, err)
			os.Exit(1)
		}

		if len(listeners) == 0 {
			fmt.Fprintf(os.Stderr, "no processes found listening on %s\n", arg)
			hasError = true
			continue
		}

		seen := make(map[int]bool)
		for _, l := range listeners {
			if seen[l.PID] {
				continue
			}
			seen[l.PID] = true

			ctx, err := process.GatherContext(l.PID, l.Port)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not get info for PID %d: %v\n", l.PID, err)
				continue
			}

			strategy := kill.RecommendedStrategy(ctx)
			action := kill.Action{
				Strategy: strategy,
				Context:  ctx,
				Force:    opts.force,
			}

			desc := kill.Describe(action)
			contextInfo := formatContext(ctx, l)

			fmt.Printf("[dry-run] %s%s\n", desc, contextInfo)
			if len(ctx.Info.Children) > 0 {
				fmt.Printf("  child PIDs: %v\n", ctx.Info.Children)
			}
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func formatContext(ctx process.Context, l port.Listener) string {
	parts := []string{
		fmt.Sprintf(" (PID %d, port %d/%s", ctx.Info.PID, l.Port, l.Protocol),
	}
	if ctx.Info.Command != "" {
		cmd := ctx.Info.Command
		if len(cmd) > 40 {
			cmd = cmd[:37] + "..."
		}
		parts = append(parts, fmt.Sprintf(", %s", cmd))
	}
	if ctx.IsContainerized() {
		name := ctx.Container.Name
		if name == "" {
			name = container.ShortID(ctx.Container.ID)
		}
		parts = append(parts, fmt.Sprintf(", %s container %s", ctx.Container.Runtime, name))
	}
	if ctx.IsSystemdManaged() {
		parts = append(parts, fmt.Sprintf(", systemd %s", ctx.SystemdUnit))
	}
	return strings.Join(parts, "") + ")"
}

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dnl/zap/internal/port"
	"github.com/dnl/zap/internal/process"
)

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
		fmt.Println("zap v0.1.0")
		os.Exit(0)
	}

	if len(opts.ports) == 0 {
		listeners, err := port.DetectAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error detecting ports: %v\n", err)
			os.Exit(1)
		}
		if len(listeners) == 0 {
			fmt.Println("No listening ports found.")
			os.Exit(0)
		}
		printListeners(listeners)
		os.Exit(0)
	}

	for _, arg := range opts.ports {
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
			os.Exit(1)
		}

		for _, l := range listeners {
			info, err := process.Gather(l.PID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not get info for PID %d: %v\n", l.PID, err)
				continue
			}

			if opts.dryRun {
				fmt.Printf("[dry-run] would kill PID %d (%s) on port %d/%s\n",
					info.PID, info.Command, l.Port, l.Protocol)
				if len(info.Children) > 0 {
					fmt.Printf("  child PIDs: %v\n", info.Children)
				}
				continue
			}

			sig := "SIGTERM"
			if opts.force {
				sig = "SIGKILL"
			}
			fmt.Printf("killing PID %d (%s) on port %d/%s with %s\n",
				info.PID, info.Command, l.Port, l.Protocol, sig)

			_ = opts.force // TODO: use actual signal in kill phase
		}
	}
}

func printListeners(listeners []port.Listener) {
	type portGroup struct {
		port     int
		protocol string
		pids     []int
	}

	groups := make(map[string]*portGroup)
	for _, l := range listeners {
		key := fmt.Sprintf("%d/%s", l.Port, l.Protocol)
		if g, ok := groups[key]; ok {
			g.pids = append(g.pids, l.PID)
		} else {
			groups[key] = &portGroup{
				port:     l.Port,
				protocol: l.Protocol,
				pids:     []int{l.PID},
			}
		}
	}

	var sorted []*portGroup
	for _, g := range groups {
		sorted = append(sorted, g)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].port < sorted[j].port
	})

	for _, g := range sorted {
		var cmds []string
		for _, pid := range g.pids {
			info, err := process.Gather(pid)
			cmd := fmt.Sprintf("PID %d", pid)
			if err == nil && info.Command != "" {
				c := info.Command
				if len(c) > 60 {
					c = c[:57] + "..."
				}
				cmd = fmt.Sprintf("PID %d (%s)", pid, c)
			}
			cmds = append(cmds, cmd)
		}
		fmt.Printf(":%d/%s  %s\n", g.port, g.protocol, strings.Join(cmds, ", "))
	}
}

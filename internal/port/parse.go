package port

import (
	"fmt"
	"strconv"
	"strings"
)

// Query represents a parsed port query from user input.
type Query struct {
	Interface string // e.g. "0.0.0.0", "localhost", "" for any
	StartPort int
	EndPort   int // same as StartPort for single port
}

// IsSinglePort returns true if this query targets a single port.
func (q Query) IsSinglePort() bool {
	return q.StartPort == q.EndPort
}

// Contains returns true if the given port falls within this query's range.
func (q Query) Contains(port int) bool {
	return port >= q.StartPort && port <= q.EndPort
}

// Parse parses a port argument string into a Query.
// Supported formats:
//   - ":3000"          → any interface, port 3000
//   - "3000"           → any interface, port 3000
//   - ":8080-8090"     → any interface, port range 8080-8090
//   - "localhost:5432" → localhost, port 5432
//   - "0.0.0.0:80"    → all interfaces, port 80
func Parse(arg string) (Query, error) {
	if arg == "" {
		return Query{}, fmt.Errorf("empty port argument")
	}

	var iface, portPart string

	// Check if there's an interface prefix
	if idx := strings.LastIndex(arg, ":"); idx >= 0 {
		prefix := arg[:idx]
		portPart = arg[idx+1:]
		// Only treat as interface if prefix is non-empty and not just another port number
		if prefix != "" && !isNumeric(prefix) {
			iface = prefix
		}
	} else {
		// No colon at all, treat entire arg as port
		portPart = arg
	}

	if portPart == "" {
		return Query{}, fmt.Errorf("missing port number in %q", arg)
	}

	// Parse port range
	startPort, endPort, err := parsePortRange(portPart)
	if err != nil {
		return Query{}, fmt.Errorf("invalid port in %q: %w", arg, err)
	}

	return Query{
		Interface: iface,
		StartPort: startPort,
		EndPort:   endPort,
	}, nil
}

func parsePortRange(s string) (start, end int, err error) {
	if idx := strings.Index(s, "-"); idx >= 0 {
		start, err = parsePort(s[:idx])
		if err != nil {
			return 0, 0, err
		}
		end, err = parsePort(s[idx+1:])
		if err != nil {
			return 0, 0, err
		}
		if start > end {
			return 0, 0, fmt.Errorf("invalid range: start port %d > end port %d", start, end)
		}
		return start, end, nil
	}

	p, err := parsePort(s)
	if err != nil {
		return 0, 0, err
	}
	return p, p, nil
}

func parsePort(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid port number", s)
	}
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("port %d out of range (1-65535)", p)
	}
	return p, nil
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

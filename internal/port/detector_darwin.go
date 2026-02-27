//go:build darwin

package port

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Detect finds all processes listening on ports matching the query.
func Detect(q Query) ([]Listener, error) {
	return detectWithLSOF(q)
}

// DetectAll finds all listening processes across all ports.
func DetectAll() ([]Listener, error) {
	return detectWithLSOF(Query{StartPort: 1, EndPort: 65535})
}

func detectWithLSOF(q Query) ([]Listener, error) {
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P", "-F", "pcn")
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		// lsof exits 1 when no files match; empty output means truly nothing
		return nil, nil
	}
	return parseLSOFOutput(out, q)
}

func parseLSOFOutput(data []byte, q Query) ([]Listener, error) {
	var listeners []Listener
	seen := make(map[string]bool)

	var pid int
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		field := line[0]
		value := line[1:]

		switch field {
		case 'p':
			p, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			pid = p
		case 'n':
			// Format: *:3000 or 127.0.0.1:3000 or [::]:3000 or [::1]:3000
			colonIdx := strings.LastIndex(value, ":")
			if colonIdx < 0 {
				continue
			}
			portStr := value[colonIdx+1:]
			p, err := strconv.Atoi(portStr)
			if err != nil {
				continue
			}
			if !q.Contains(p) {
				continue
			}

			iface := value[:colonIdx]
			if iface == "*" || iface == "" {
				iface = "0.0.0.0"
			} else {
				iface = strings.Trim(iface, "[]")
			}

			if q.Interface != "" && iface != q.Interface && iface != "0.0.0.0" && iface != "::" {
				continue
			}

			key := fmt.Sprintf("%d:%d", pid, p)
			if seen[key] {
				continue
			}
			seen[key] = true

			listeners = append(listeners, Listener{
				PID:       pid,
				Port:      p,
				Protocol:  "tcp",
				Interface: iface,
			})
		}
	}
	return listeners, scanner.Err()
}

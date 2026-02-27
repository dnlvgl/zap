//go:build darwin

package container

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// Detect checks if the port is being served by a container on macOS.
// On macOS, Docker and Podman run in a VM so cgroup-based detection doesn't
// work. Instead, we query the container runtime directly by port.
func Detect(pid, port int) *Info {
	return detectByPort(port)
}

type containerPSEntry struct {
	ID    string `json:"ID"`
	Names string `json:"Names"`
	Ports string `json:"Ports"`
}

func detectByPort(port int) *Info {
	portStr := strconv.Itoa(port)
	for _, runtime := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(runtime); err != nil {
			continue
		}
		cmd := exec.Command(runtime, "ps", "--format", "{{json .}}")
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			var entry containerPSEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			// Ports format: "0.0.0.0:3000->3000/tcp, :::3000->3000/tcp"
			if strings.Contains(entry.Ports, ":"+portStr+"->") {
				name := strings.TrimPrefix(entry.Names, "/")
				return &Info{
					ID:      entry.ID,
					Name:    name,
					Runtime: runtime,
				}
			}
		}
	}
	return nil
}

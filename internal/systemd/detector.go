package systemd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Detect checks if a process is managed by systemd and returns the unit name.
// Returns empty string if not a systemd-managed process.
func Detect(pid int) string {
	// First try reading cgroup for systemd slice info
	unit := detectFromCgroup(pid)
	if unit != "" {
		return unit
	}
	// Fallback: ask systemctl
	return detectFromSystemctl(pid)
}

func detectFromCgroup(pid int) string {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cgroup"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		// Format: hierarchy-ID:controller-list:cgroup-path
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		cgroupPath := parts[2]

		// Look for .service in the path
		// e.g., /system.slice/nginx.service
		// e.g., /system.slice/docker.service
		segments := strings.Split(cgroupPath, "/")
		for _, seg := range segments {
			if strings.HasSuffix(seg, ".service") {
				// Skip generic/infrastructure services
			if seg == "docker.service" || seg == "podman.service" || seg == "containerd.service" {
				continue
			}
			// Skip user session slices (user@1000.service etc.)
			if strings.HasPrefix(seg, "user@") {
				continue
			}
				return seg
			}
		}
	}

	return ""
}

func detectFromSystemctl(pid int) string {
	cmd := exec.Command("systemctl", "status", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse output for unit name
	// First line typically contains: ● unit-name.service - Description
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		// Remove the bullet character if present
		line = strings.TrimLeft(line, "● ")
		if strings.Contains(line, ".service") {
			parts := strings.Fields(line)
			if len(parts) > 0 && strings.HasSuffix(parts[0], ".service") {
				return parts[0]
			}
		}
	}

	return ""
}

// Stop stops a systemd service.
func Stop(unit string) error {
	cmd := exec.Command("systemctl", "stop", unit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsAvailable checks if systemd is running on this system.
func IsAvailable() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

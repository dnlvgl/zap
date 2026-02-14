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
	return parseCgroupUnit(string(data))
}

func parseCgroupUnit(content string) string {
	for _, line := range strings.Split(content, "\n") {
		// Format: hierarchy-ID:controller-list:cgroup-path
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		cgroupPath := parts[2]

		segments := strings.Split(cgroupPath, "/")
		for _, seg := range segments {
			if !strings.HasSuffix(seg, ".service") {
				continue
			}
			if isInfrastructureUnit(seg) {
				continue
			}
			return seg
		}
	}
	return ""
}

// isInfrastructureUnit returns true for systemd units that should never be stopped
// by zap because they manage user sessions or container infrastructure.
func isInfrastructureUnit(unit string) bool {
	// User session managers (user@1000.service etc.)
	if strings.HasPrefix(unit, "user@") {
		return true
	}
	// Container runtime services
	switch unit {
	case "docker.service", "podman.service", "containerd.service":
		return true
	}
	// Display managers and session services
	switch unit {
	case "gdm.service", "sddm.service", "lightdm.service", "display-manager.service":
		return true
	}
	return false
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
				unit := parts[0]
				if isInfrastructureUnit(unit) {
					continue
				}
				if !isMainPIDOfUnit(pid, unit) {
					continue
				}
				return unit
			}
		}
	}

	return ""
}

// isMainPIDOfUnit checks whether the given PID is the main process of a systemd unit,
// not just a descendant running inside its cgroup.
func isMainPIDOfUnit(pid int, unit string) bool {
	cmd := exec.Command("systemctl", "show", "--property=MainPID", unit)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	// Output is like "MainPID=12345"
	s := strings.TrimSpace(string(out))
	mainPID := strings.TrimPrefix(s, "MainPID=")
	return mainPID == strconv.Itoa(pid)
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

package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Info holds container details for a process.
type Info struct {
	ID      string
	Name    string
	Runtime string // "podman" or "docker"
}

// Detect checks if a process is running inside a container.
// Returns nil if the process is not containerized.
func Detect(pid int) *Info {
	cgroupPath := filepath.Join("/proc", strconv.Itoa(pid), "cgroup")
	data, err := os.ReadFile(cgroupPath)
	if err != nil {
		return nil
	}

	containerID, runtime := parseCgroup(string(data))
	if containerID == "" {
		return nil
	}

	name := getContainerName(containerID, runtime)

	return &Info{
		ID:      containerID,
		Name:    name,
		Runtime: runtime,
	}
}

// libpod-<id>.scope for Podman, docker-<id>.scope for Docker
var (
	libpodRe = regexp.MustCompile(`libpod-([0-9a-f]{64})`)
	dockerRe = regexp.MustCompile(`docker-([0-9a-f]{64})`)
	// Also match /docker/<id> or /lxc/<id> style paths
	slashDockerRe = regexp.MustCompile(`/docker/([0-9a-f]{64})`)
	slashLXCRe    = regexp.MustCompile(`/lxc/([0-9a-f]{64})`)
)

func parseCgroup(content string) (containerID, runtime string) {
	for _, line := range strings.Split(content, "\n") {
		// Podman (libpod)
		if m := libpodRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], detectRuntime("podman")
		}
		// Docker scope style
		if m := dockerRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], detectRuntime("docker")
		}
		// Docker slash style (/docker/<id>)
		if m := slashDockerRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], detectRuntime("docker")
		}
		// LXC style
		if m := slashLXCRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], "docker" // LXC-based docker
		}
	}
	return "", ""
}

// detectRuntime verifies which runtime is actually available.
func detectRuntime(hint string) string {
	if hint == "podman" {
		if _, err := exec.LookPath("podman"); err == nil {
			return "podman"
		}
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker"
	}
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	return hint
}

func getContainerName(containerID, runtime string) string {
	cmd := exec.Command(runtime, "inspect", "--format", "{{.Name}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		// Try short ID (first 12 chars)
		if len(containerID) > 12 {
			cmd = exec.Command(runtime, "inspect", "--format", "{{.Name}}", containerID[:12])
			out, err = cmd.Output()
			if err != nil {
				return ""
			}
		} else {
			return ""
		}
	}
	name := strings.TrimSpace(string(out))
	// Docker prefixes names with /
	name = strings.TrimPrefix(name, "/")
	return name
}

// Stop stops a container gracefully.
func Stop(containerID, runtime string) error {
	cmd := exec.Command(runtime, "stop", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Kill forcefully kills a container.
func Kill(containerID, runtime string) error {
	cmd := exec.Command(runtime, "kill", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ShortID returns the first 12 characters of a container ID.
func ShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// String returns a human-readable description of the container.
func (i Info) String() string {
	name := i.Name
	if name == "" {
		name = ShortID(i.ID)
	}
	return fmt.Sprintf("%s container %s", i.Runtime, name)
}

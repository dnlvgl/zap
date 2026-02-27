//go:build linux

package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// libpod-<id>.scope for Podman, docker-<id>.scope for Docker
var (
	libpodRe      = regexp.MustCompile(`libpod-([0-9a-f]{64})`)
	dockerRe      = regexp.MustCompile(`docker-([0-9a-f]{64})`)
	slashDockerRe = regexp.MustCompile(`/docker/([0-9a-f]{64})`)
	slashLXCRe    = regexp.MustCompile(`/lxc/([0-9a-f]{64})`)
)

// Detect checks if a process is running inside a container.
// Returns nil if the process is not containerized.
func Detect(pid, port int) *Info {
	cgroupPath := filepath.Join("/proc", strconv.Itoa(pid), "cgroup")
	data, err := os.ReadFile(cgroupPath)
	if err != nil {
		return nil
	}

	containerID, runtimeHint := parseCgroup(string(data))
	if containerID == "" {
		return nil
	}

	runtime := detectRuntime(runtimeHint)
	name := getContainerName(containerID, runtime)

	return &Info{
		ID:      containerID,
		Name:    name,
		Runtime: runtime,
	}
}

func parseCgroup(content string) (containerID, runtime string) {
	for _, line := range strings.Split(content, "\n") {
		// Podman (libpod)
		if m := libpodRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], "podman"
		}
		// Docker scope style
		if m := dockerRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], "docker"
		}
		// Docker slash style (/docker/<id>)
		if m := slashDockerRe.FindStringSubmatch(line); len(m) > 1 {
			return m[1], "docker"
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

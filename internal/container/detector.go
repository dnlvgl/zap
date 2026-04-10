package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const detectionTimeout = 5 * time.Second
const operationTimeout = 30 * time.Second

// Info holds container details for a process.
type Info struct {
	ID      string
	Name    string
	Runtime string // "podman" or "docker"
}

// Stop stops a container gracefully.
func Stop(containerID, runtime string) error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, runtime, "stop", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Kill forcefully kills a container.
func Kill(containerID, runtime string) error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, runtime, "kill", containerID)
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

func getContainerName(containerID, runtime string) string {
	ctx, cancel := context.WithTimeout(context.Background(), detectionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, runtime, "inspect", "--format", "{{.Name}}", containerID)
	out, err := cmd.Output()
	if err != nil {
		// Try short ID (first 12 chars)
		if len(containerID) > 12 {
			ctx2, cancel2 := context.WithTimeout(context.Background(), detectionTimeout)
			defer cancel2()
			cmd = exec.CommandContext(ctx2, runtime, "inspect", "--format", "{{.Name}}", containerID[:12])
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

package process

import (
	"os"
	"syscall"
	"time"
)

// Info holds details about a running process.
type Info struct {
	PID        int
	Command    string
	Executable string
	User       string
	UID        int
	Ports      []PortBinding
	CPUPercent float64
	MemoryKB   int64
	StartTime  time.Time
	ParentPID  int
	Children   []int
}

// PortBinding describes a port a process is listening on.
type PortBinding struct {
	Port      int
	Protocol  string
	Interface string
}

// Uptime returns the duration since the process started.
func (i Info) Uptime() time.Duration {
	if i.StartTime.IsZero() {
		return 0
	}
	return time.Since(i.StartTime)
}

// IsPrivileged returns true if killing this process requires elevated privileges.
func (i Info) IsPrivileged() bool {
	return i.UID != os.Getuid() && os.Getuid() != 0
}

// Signal sends a signal to the process.
func (i Info) Signal(sig syscall.Signal) error {
	proc, err := os.FindProcess(i.PID)
	if err != nil {
		return err
	}
	return proc.Signal(sig)
}

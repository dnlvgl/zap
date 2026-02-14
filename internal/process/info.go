package process

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
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

// Gather collects information about a process by PID.
func Gather(pid int) (Info, error) {
	procPath := filepath.Join("/proc", strconv.Itoa(pid))

	if _, err := os.Stat(procPath); err != nil {
		return Info{}, fmt.Errorf("process %d not found", pid)
	}

	info := Info{PID: pid}

	// Read command line
	if cmdline, err := os.ReadFile(filepath.Join(procPath, "cmdline")); err == nil {
		// cmdline is null-separated
		parts := strings.Split(string(cmdline), "\x00")
		var nonEmpty []string
		for _, p := range parts {
			if p != "" {
				nonEmpty = append(nonEmpty, p)
			}
		}
		info.Command = strings.Join(nonEmpty, " ")
	}

	// Read executable path
	if exe, err := os.Readlink(filepath.Join(procPath, "exe")); err == nil {
		info.Executable = exe
	}

	// Read status file for UID and PPID
	if status, err := os.ReadFile(filepath.Join(procPath, "status")); err == nil {
		for _, line := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(line, "PPid:") {
				fmt.Sscanf(strings.TrimPrefix(line, "PPid:"), "%d", &info.ParentPID)
			}
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					uid, _ := strconv.Atoi(fields[1])
					info.UID = uid
					if u, err := user.LookupId(fields[1]); err == nil {
						info.User = u.Username
					} else {
						info.User = fields[1]
					}
				}
			}
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					info.MemoryKB, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			}
		}
	}

	// Read start time from /proc/PID/stat
	info.StartTime = readStartTime(pid)

	// Find child processes
	info.Children = findChildren(pid)

	return info, nil
}

func readStartTime(pid int) time.Time {
	stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return time.Time{}
	}

	// Fields in stat are space-separated, but comm (field 2) can contain spaces
	// and is enclosed in parentheses. Find the last ')' to skip past it.
	s := string(stat)
	idx := strings.LastIndex(s, ")")
	if idx < 0 {
		return time.Time{}
	}
	fields := strings.Fields(s[idx+2:]) // skip ") "
	if len(fields) < 20 {
		return time.Time{}
	}

	// Field index 19 (from after comm) is starttime in clock ticks
	startTicks, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return time.Time{}
	}

	// Get system boot time
	bootTime := getBootTime()
	if bootTime.IsZero() {
		return time.Time{}
	}

	clkTck := uint64(100) // sysconf(_SC_CLK_TCK), typically 100 on Linux
	startSecs := startTicks / clkTck
	return bootTime.Add(time.Duration(startSecs) * time.Second)
}

func getBootTime() time.Time {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				btime, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return time.Unix(btime, 0)
				}
			}
		}
	}
	return time.Time{}
}

func findChildren(pid int) []int {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "task", strconv.Itoa(pid), "children"))
	if err != nil {
		return nil
	}
	var children []int
	for _, s := range strings.Fields(string(data)) {
		if child, err := strconv.Atoi(s); err == nil {
			children = append(children, child)
		}
	}
	return children
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

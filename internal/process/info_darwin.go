//go:build darwin

package process

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Gather collects information about a process by PID.
func Gather(pid int) (Info, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return Info{}, fmt.Errorf("process %d not found", pid)
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return Info{}, fmt.Errorf("process %d not found", pid)
	}

	info := Info{PID: pid}

	// ps -p <pid> -o ppid=,uid=,rss=,command=
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "ppid=,uid=,rss=,command=")
	out, err := cmd.Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			if ppid, err := strconv.Atoi(fields[0]); err == nil {
				info.ParentPID = ppid
			}
			if uid, err := strconv.Atoi(fields[1]); err == nil {
				info.UID = uid
				if u, err := user.LookupId(fields[1]); err == nil {
					info.User = u.Username
				} else {
					info.User = fields[1]
				}
			}
			if rss, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				info.MemoryKB = rss
			}
			info.Command = strings.Join(fields[3:], " ")
		}
	}

	info.StartTime = readStartTime(pid)
	info.Children = findChildren(pid)

	return info, nil
}

func readStartTime(pid int) time.Time {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}
	}
	s := strings.TrimSpace(string(out))
	// lstart format: "Thu Feb 27 10:30:00 2026" (space-padded single-digit days)
	t, err := time.ParseInLocation("Mon Jan _2 15:04:05 2006", s, time.Local)
	if err != nil {
		return time.Time{}
	}
	return t
}

func findChildren(pid int) []int {
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var children []int
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		childPID, err1 := strconv.Atoi(fields[0])
		parentPID, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		if parentPID == pid {
			children = append(children, childPID)
		}
	}
	return children
}

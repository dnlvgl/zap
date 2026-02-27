//go:build linux

package port

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type socketInfo struct {
	port     int
	protocol string
	iface    string
}

// Detect finds all processes listening on ports matching the query.
func Detect(q Query) ([]Listener, error) {
	return detectFromProc(q)
}

// DetectAll finds all listening processes across all ports.
func DetectAll() ([]Listener, error) {
	return detectFromProc(Query{StartPort: 1, EndPort: 65535})
}

func detectFromProc(q Query) ([]Listener, error) {
	inodeMap := make(map[uint64]socketInfo)

	for _, proto := range []string{"tcp", "tcp6", "udp", "udp6"} {
		path := filepath.Join("/proc/net", proto)
		entries, err := parseProcNet(path)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.state != 0x0A && !strings.HasPrefix(proto, "udp") {
				continue // only LISTEN state for TCP
			}
			if !q.Contains(e.localPort) {
				continue
			}
			if q.Interface != "" && e.localAddr != q.Interface && e.localAddr != "0.0.0.0" && e.localAddr != "::" {
				continue
			}
			inodeMap[e.inode] = socketInfo{
				port:     e.localPort,
				protocol: proto,
				iface:    e.localAddr,
			}
		}
	}

	if len(inodeMap) == 0 {
		return nil, nil
	}

	listeners := findPIDsForInodes(inodeMap)
	return listeners, nil
}

type procNetEntry struct {
	localAddr string
	localPort int
	state     int
	inode     uint64
}

func parseProcNet(path string) ([]procNetEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []procNetEntry
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		localAddr, localPort, err := parseHexAddr(fields[1])
		if err != nil {
			continue
		}

		state, err := strconv.ParseInt(fields[3], 16, 32)
		if err != nil {
			continue
		}

		inode, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}

		entries = append(entries, procNetEntry{
			localAddr: localAddr,
			localPort: localPort,
			state:     int(state),
			inode:     inode,
		})
	}

	return entries, scanner.Err()
}

func parseHexAddr(s string) (addr string, port int, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address format: %s", s)
	}

	p, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return "", 0, err
	}
	port = int(p)

	hexAddr := parts[0]
	switch len(hexAddr) {
	case 8: // IPv4
		b, err := hex.DecodeString(hexAddr)
		if err != nil {
			return "", 0, err
		}
		// /proc/net/tcp stores addresses in little-endian
		addr = fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0])
	case 32: // IPv6
		if hexAddr == "00000000000000000000000000000000" {
			addr = "::"
		} else {
			addr = "::" // simplified
		}
	default:
		addr = hexAddr
	}

	return addr, port, nil
}

func findPIDsForInodes(inodeMap map[uint64]socketInfo) []Listener {
	var listeners []Listener
	seen := make(map[string]bool)

	procDir, err := os.Open("/proc")
	if err != nil {
		return nil
	}
	defer procDir.Close()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry, "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}

			if !strings.HasPrefix(link, "socket:[") {
				continue
			}

			inodeStr := link[8 : len(link)-1]
			inode, err := strconv.ParseUint(inodeStr, 10, 64)
			if err != nil {
				continue
			}

			info, ok := inodeMap[inode]
			if !ok {
				continue
			}

			key := fmt.Sprintf("%d:%d:%s", pid, info.port, info.protocol)
			if seen[key] {
				continue
			}
			seen[key] = true

			listeners = append(listeners, Listener{
				PID:       pid,
				Port:      info.port,
				Protocol:  info.protocol,
				Interface: info.iface,
			})
		}
	}

	return listeners
}

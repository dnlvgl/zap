package port

// Listener represents a process listening on a port.
type Listener struct {
	PID       int
	Port      int
	Protocol  string // "tcp", "tcp6", "udp", "udp6"
	Interface string // parsed from local address
}

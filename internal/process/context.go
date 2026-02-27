package process

import (
	"github.com/dnlvgl/zap/internal/container"
	"github.com/dnlvgl/zap/internal/systemd"
)

// Context combines process info with container and systemd detection.
type Context struct {
	Info        Info
	Container   *container.Info
	SystemdUnit string
}

// GatherContext collects full process context including container and systemd info.
func GatherContext(pid, port int) (Context, error) {
	info, err := Gather(pid)
	if err != nil {
		return Context{}, err
	}

	ctx := Context{
		Info:        info,
		Container:   container.Detect(pid, port),
		SystemdUnit: systemd.Detect(pid),
	}

	return ctx, nil
}

// IsContainerized returns true if the process runs inside a container.
func (c Context) IsContainerized() bool {
	return c.Container != nil
}

// IsSystemdManaged returns true if the process is managed by a systemd unit.
func (c Context) IsSystemdManaged() bool {
	return c.SystemdUnit != ""
}

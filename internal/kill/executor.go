package kill

import (
	"fmt"
	"syscall"

	"github.com/dnl/zap/internal/container"
	"github.com/dnl/zap/internal/process"
	"github.com/dnl/zap/internal/systemd"
)

// Strategy describes how to stop a process.
type Strategy int

const (
	StrategySignal    Strategy = iota // Send SIGTERM or SIGKILL
	StrategyContainer                 // podman/docker stop or kill
	StrategySystemd                   // systemctl stop
)

func (s Strategy) String() string {
	switch s {
	case StrategySignal:
		return "signal"
	case StrategyContainer:
		return "container"
	case StrategySystemd:
		return "systemd"
	default:
		return "unknown"
	}
}

// Action describes a kill action to take.
type Action struct {
	Strategy Strategy
	Context  process.Context
	Force    bool
}

// RecommendedStrategy picks the best strategy for a given process context.
func RecommendedStrategy(ctx process.Context) Strategy {
	if ctx.IsContainerized() {
		return StrategyContainer
	}
	if ctx.IsSystemdManaged() {
		return StrategySystemd
	}
	return StrategySignal
}

// AvailableStrategies returns all applicable strategies for a process context.
func AvailableStrategies(ctx process.Context) []Strategy {
	var strategies []Strategy
	if ctx.IsContainerized() {
		strategies = append(strategies, StrategyContainer)
	}
	if ctx.IsSystemdManaged() {
		strategies = append(strategies, StrategySystemd)
	}
	strategies = append(strategies, StrategySignal)
	return strategies
}

// Execute performs the kill action.
func Execute(action Action) error {
	switch action.Strategy {
	case StrategyContainer:
		return executeContainer(action)
	case StrategySystemd:
		return executeSystemd(action)
	case StrategySignal:
		return executeSignal(action)
	default:
		return fmt.Errorf("unknown strategy: %v", action.Strategy)
	}
}

// Describe returns a human-readable description of what the action will do.
func Describe(action Action) string {
	switch action.Strategy {
	case StrategyContainer:
		verb := "stop"
		if action.Force {
			verb = "kill"
		}
		c := action.Context.Container
		name := c.Name
		if name == "" {
			name = container.ShortID(c.ID)
		}
		return fmt.Sprintf("%s %s %s", c.Runtime, verb, name)
	case StrategySystemd:
		return fmt.Sprintf("systemctl stop %s", action.Context.SystemdUnit)
	case StrategySignal:
		sig := "SIGTERM"
		if action.Force {
			sig = "SIGKILL"
		}
		return fmt.Sprintf("kill -%s %d", sig, action.Context.Info.PID)
	default:
		return "unknown action"
	}
}

func executeContainer(action Action) error {
	c := action.Context.Container
	if action.Force {
		return container.Kill(c.ID, c.Runtime)
	}
	return container.Stop(c.ID, c.Runtime)
}

func executeSystemd(action Action) error {
	return systemd.Stop(action.Context.SystemdUnit)
}

func executeSignal(action Action) error {
	sig := syscall.SIGTERM
	if action.Force {
		sig = syscall.SIGKILL
	}
	return action.Context.Info.Signal(sig)
}

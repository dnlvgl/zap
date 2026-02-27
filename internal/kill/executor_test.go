package kill

import (
	"testing"

	"github.com/dnlvgl/zap/internal/container"
	"github.com/dnlvgl/zap/internal/process"
)

func TestRecommendedStrategy(t *testing.T) {
	tests := []struct {
		name string
		ctx  process.Context
		want Strategy
	}{
		{
			name: "bare process",
			ctx:  process.Context{Info: process.Info{PID: 1234}},
			want: StrategySignal,
		},
		{
			name: "container process",
			ctx: process.Context{
				Info:      process.Info{PID: 1234},
				Container: &container.Info{ID: "abc", Runtime: "podman"},
			},
			want: StrategyContainer,
		},
		{
			name: "systemd process",
			ctx: process.Context{
				Info:        process.Info{PID: 1234},
				SystemdUnit: "nginx.service",
			},
			want: StrategySystemd,
		},
		{
			name: "container takes priority over systemd",
			ctx: process.Context{
				Info:        process.Info{PID: 1234},
				Container:   &container.Info{ID: "abc", Runtime: "docker"},
				SystemdUnit: "docker.service",
			},
			want: StrategyContainer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecommendedStrategy(tt.ctx)
			if got != tt.want {
				t.Errorf("RecommendedStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvailableStrategies(t *testing.T) {
	ctx := process.Context{
		Info:        process.Info{PID: 1234},
		Container:   &container.Info{ID: "abc", Runtime: "podman"},
		SystemdUnit: "myapp.service",
	}

	strategies := AvailableStrategies(ctx)
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d", len(strategies))
	}
	if strategies[0] != StrategyContainer {
		t.Errorf("first strategy = %v, want container", strategies[0])
	}
	if strategies[1] != StrategySystemd {
		t.Errorf("second strategy = %v, want systemd", strategies[1])
	}
	if strategies[2] != StrategySignal {
		t.Errorf("third strategy = %v, want signal", strategies[2])
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{
			name: "signal SIGTERM",
			action: Action{
				Strategy: StrategySignal,
				Context:  process.Context{Info: process.Info{PID: 1234}},
			},
			want: "kill -SIGTERM 1234",
		},
		{
			name: "signal SIGKILL",
			action: Action{
				Strategy: StrategySignal,
				Context:  process.Context{Info: process.Info{PID: 1234}},
				Force:    true,
			},
			want: "kill -SIGKILL 1234",
		},
		{
			name: "container stop",
			action: Action{
				Strategy: StrategyContainer,
				Context: process.Context{
					Container: &container.Info{ID: "abc123def456", Name: "myapp", Runtime: "podman"},
				},
			},
			want: "podman stop myapp",
		},
		{
			name: "container kill",
			action: Action{
				Strategy: StrategyContainer,
				Context: process.Context{
					Container: &container.Info{ID: "abc123def456", Name: "myapp", Runtime: "docker"},
				},
				Force: true,
			},
			want: "docker kill myapp",
		},
		{
			name: "systemd stop",
			action: Action{
				Strategy: StrategySystemd,
				Context:  process.Context{SystemdUnit: "nginx.service"},
			},
			want: "systemctl stop nginx.service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Describe(tt.action)
			if got != tt.want {
				t.Errorf("Describe() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStrategyString(t *testing.T) {
	if s := StrategySignal.String(); s != "signal" {
		t.Errorf("StrategySignal.String() = %q", s)
	}
	if s := StrategyContainer.String(); s != "container" {
		t.Errorf("StrategyContainer.String() = %q", s)
	}
	if s := StrategySystemd.String(); s != "systemd" {
		t.Errorf("StrategySystemd.String() = %q", s)
	}
}

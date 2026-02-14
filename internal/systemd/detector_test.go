package systemd

import (
	"testing"
)

func TestIsInfrastructureUnit(t *testing.T) {
	tests := []struct {
		unit string
		want bool
	}{
		{"user@1000.service", true},
		{"user@0.service", true},
		{"docker.service", true},
		{"podman.service", true},
		{"containerd.service", true},
		{"gdm.service", true},
		{"sddm.service", true},
		{"lightdm.service", true},
		{"display-manager.service", true},
		{"nginx.service", false},
		{"sshd.service", false},
		{"myapp.service", false},
	}

	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			if got := isInfrastructureUnit(tt.unit); got != tt.want {
				t.Errorf("isInfrastructureUnit(%q) = %v, want %v", tt.unit, got, tt.want)
			}
		})
	}
}

func TestParseCgroupUnit(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "nginx service",
			content: "0::/system.slice/nginx.service",
			want:    "nginx.service",
		},
		{
			name:    "syncthing service",
			content: "0::/user.slice/user-1000.slice/user@1000.service/app.slice/syncthing.service",
			want:    "syncthing.service",
		},
		{
			name:    "skip docker runtime",
			content: "0::/system.slice/docker.service",
			want:    "",
		},
		{
			name:    "skip podman runtime",
			content: "0::/system.slice/podman.service",
			want:    "",
		},
		{
			name:    "skip containerd runtime",
			content: "0::/system.slice/containerd.service",
			want:    "",
		},
		{
			name:    "skip user session",
			content: "0::/user.slice/user-1000.slice/user@1000.service",
			want:    "",
		},
		{
			name:    "bare process",
			content: "0::/user.slice/user-1000.slice/session-1.scope",
			want:    "",
		},
		{
			name:    "empty",
			content: "",
			want:    "",
		},
		{
			name: "multiline with service",
			content: "12:pids:/user.slice/user-1000.slice\n" +
				"0::/system.slice/sshd.service",
			want: "sshd.service",
		},
		{
			name:    "v1 cgroup style",
			content: "1:name=systemd:/system.slice/postgresql.service",
			want:    "postgresql.service",
		},
		{
			name:    "skip gdm service",
			content: "0::/system.slice/gdm.service",
			want:    "",
		},
		{
			name:    "skip display-manager service",
			content: "0::/system.slice/display-manager.service",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCgroupUnit(tt.content)
			if got != tt.want {
				t.Errorf("parseCgroupUnit() = %q, want %q", got, tt.want)
			}
		})
	}
}

//go:build linux

package container

import "testing"

func TestParseCgroup(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantID   string
		wantHint string
	}{
		{
			name:     "podman libpod scope",
			content:  "0::/system.slice/libpod-" + testID + ".scope",
			wantID:   testID,
			wantHint: "podman",
		},
		{
			name:     "docker scope",
			content:  "0::/system.slice/docker-" + testID + ".scope",
			wantID:   testID,
			wantHint: "docker",
		},
		{
			name:     "docker slash style",
			content:  "12:memory:/docker/" + testID,
			wantID:   testID,
			wantHint: "docker",
		},
		{
			name:     "lxc style",
			content:  "12:memory:/lxc/" + testID,
			wantID:   testID,
			wantHint: "docker",
		},
		{
			name:    "bare process",
			content: "0::/user.slice/user-1000.slice/session-1.scope",
			wantID:  "",
		},
		{
			name:    "empty",
			content: "",
			wantID:  "",
		},
		{
			name: "multiline with podman",
			content: "12:pids:/user.slice/user-1000.slice\n" +
				"0::/system.slice/libpod-" + testID + ".scope",
			wantID:   testID,
			wantHint: "podman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, hint := parseCgroup(tt.content)
			if id != tt.wantID {
				t.Errorf("containerID = %q, want %q", id, tt.wantID)
			}
			if tt.wantID != "" && hint != tt.wantHint {
				t.Errorf("runtime hint = %q, want %q", hint, tt.wantHint)
			}
		})
	}
}

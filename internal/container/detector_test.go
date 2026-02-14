package container

import "testing"

const testID = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

func TestParseCgroup(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantID    string
		wantHint  string
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

func TestShortID(t *testing.T) {
	if got := ShortID(testID); got != testID[:12] {
		t.Errorf("ShortID = %q, want %q", got, testID[:12])
	}
	if got := ShortID("abc"); got != "abc" {
		t.Errorf("ShortID short = %q, want %q", got, "abc")
	}
}

func TestInfoString(t *testing.T) {
	info := Info{ID: testID, Name: "myapp", Runtime: "podman"}
	if got := info.String(); got != "podman container myapp" {
		t.Errorf("String = %q", got)
	}

	noName := Info{ID: testID, Runtime: "docker"}
	if got := noName.String(); got != "docker container "+testID[:12] {
		t.Errorf("String no name = %q", got)
	}
}

package container

import "testing"

const testID = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"


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

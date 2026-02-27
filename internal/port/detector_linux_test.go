//go:build linux

package port

import "testing"

func TestParseHexAddr(t *testing.T) {
	tests := []struct {
		input    string
		wantAddr string
		wantPort int
		wantErr  bool
	}{
		{"0100007F:0BB8", "127.0.0.1", 3000, false},
		{"00000000:1F90", "0.0.0.0", 8080, false},
		{"0100007F:0050", "127.0.0.1", 80, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			addr, port, err := parseHexAddr(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if addr != tt.wantAddr {
				t.Errorf("addr = %q, want %q", addr, tt.wantAddr)
			}
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
		})
	}
}

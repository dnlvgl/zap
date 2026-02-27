package port

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		wantIface string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		// Single ports
		{":3000", "", 3000, 3000, false},
		{"3000", "", 3000, 3000, false},
		{":80", "", 80, 80, false},
		{":65535", "", 65535, 65535, false},

		// Port ranges
		{":8080-8090", "", 8080, 8090, false},
		{"8080-8090", "", 8080, 8090, false},

		// With interface
		{"localhost:5432", "localhost", 5432, 5432, false},
		{"0.0.0.0:80", "0.0.0.0", 80, 80, false},
		{"127.0.0.1:3000", "127.0.0.1", 3000, 3000, false},

		// Errors
		{"", "", 0, 0, true},
		{":", "", 0, 0, true},
		{":0", "", 0, 0, true},
		{":65536", "", 0, 0, true},
		{":abc", "", 0, 0, true},
		{":9000-8000", "", 0, 0, true}, // reversed range
		{"localhost:", "", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			q, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got %+v", tt.input, q)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if q.Interface != tt.wantIface {
				t.Errorf("Parse(%q).Interface = %q, want %q", tt.input, q.Interface, tt.wantIface)
			}
			if q.StartPort != tt.wantStart {
				t.Errorf("Parse(%q).StartPort = %d, want %d", tt.input, q.StartPort, tt.wantStart)
			}
			if q.EndPort != tt.wantEnd {
				t.Errorf("Parse(%q).EndPort = %d, want %d", tt.input, q.EndPort, tt.wantEnd)
			}
		})
	}
}

func TestQueryContains(t *testing.T) {
	q := Query{StartPort: 8080, EndPort: 8090}

	if !q.Contains(8080) {
		t.Error("expected 8080 to be in range")
	}
	if !q.Contains(8085) {
		t.Error("expected 8085 to be in range")
	}
	if !q.Contains(8090) {
		t.Error("expected 8090 to be in range")
	}
	if q.Contains(8079) {
		t.Error("expected 8079 to NOT be in range")
	}
	if q.Contains(8091) {
		t.Error("expected 8091 to NOT be in range")
	}
}

func TestQueryIsSinglePort(t *testing.T) {
	single := Query{StartPort: 3000, EndPort: 3000}
	if !single.IsSinglePort() {
		t.Error("expected single port query")
	}

	ranged := Query{StartPort: 3000, EndPort: 3010}
	if ranged.IsSinglePort() {
		t.Error("expected range query")
	}
}


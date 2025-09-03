package context

import (
	"go-cache/policies"
	"testing"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		name    string
		cap     int
		wantErr bool
	}{
		{"0 cap", 0, false},
		{"-1 cap", -1, true},
		{"1 cap", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := NewOptions(tt.cap, policies.NewFIFO[int]())
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected err=nil, got %v", err)
				}
				if opts.Capacity != tt.cap {
					t.Errorf("expected cap=%d, got %d", tt.cap, opts.Capacity)
				}
			}
		})
	}

}

func TestOptions_Equals(t *testing.T) {
	opts1, _ := NewOptions(1, policies.NewFIFO[int]())
	opts2, _ := NewOptions(1, policies.NewFIFO[int]())
	opts3, _ := NewOptions(2, policies.NewFIFO[int]())

	if !opts1.Equals(opts2) {
		t.Errorf("expected opts1 to equal opts2")
	}
	if opts1.Equals(opts3) {
		t.Errorf("expected opts1 to not equal opts3")
	}
}

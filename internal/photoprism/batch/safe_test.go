package batch

import "testing"

func TestIntToSafeUint(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		fallback uint
		expect   uint
	}{
		{name: "Negative", value: -5, fallback: 42, expect: 42},
		{name: "Zero", value: 0, fallback: 7, expect: 0},
		{name: "Positive", value: 15, fallback: 9, expect: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intToSafeUint(tt.value, tt.fallback); got != tt.expect {
				t.Fatalf("expected %d, got %d", tt.expect, got)
			}
		})
	}
}

package watchdog

import (
	"testing"
)

func TestWraparoundStatus(t *testing.T) {
	tests := []struct {
		currentAge int64
		maxAge     int64
		expected   string
	}{
		{20000000, 200000000, "HEALTHY"},  // 10%
		{110000000, 200000000, "WARNING"},  // 55%
		{160000000, 200000000, "CRITICAL"}, // 80%
	}

	for _, tt := range tests {
		percent := (float64(tt.currentAge) / float64(tt.maxAge)) * 100.0
		status := "HEALTHY"
		if percent >= 75.0 || tt.currentAge >= 150000000 {
			status = "CRITICAL"
		} else if percent >= 50.0 || tt.currentAge >= 100000000 {
			status = "WARNING"
		}

		if status != tt.expected {
			t.Errorf("currentAge=%d, maxAge=%d: got status %s, want %s", tt.currentAge, tt.maxAge, status, tt.expected)
		}
	}
}

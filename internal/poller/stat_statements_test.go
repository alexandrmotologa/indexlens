package poller

import (
	"testing"
)

func TestParseQueryID(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"123456789", 123456789},
		{"-9223372036854775808", -9223372036854775808},
		{"invalid", 0},
		{"0", 0},
	}

	for _, tt := range tests {
		actual := ParseQueryID(tt.input)
		if actual != tt.expected {
			t.Errorf("ParseQueryID(%q) = %d; want %d", tt.input, actual, tt.expected)
		}
	}
}

func TestCacheHitRatioCalculation(t *testing.T) {
	tests := []struct {
		hit      int64
		read     int64
		expected float64
	}{
		{100, 0, 100.0},
		{50, 50, 50.0},
		{0, 100, 0.0},
		{0, 0, 100.0},
	}

	for _, tt := range tests {
		total := tt.hit + tt.read
		var ratio float64
		if total > 0 {
			ratio = (float64(tt.hit) / float64(total)) * 100.0
		} else {
			ratio = 100.0
		}

		if ratio != tt.expected {
			t.Errorf("CacheHitRatio(%d, %d) = %f; want %f", tt.hit, tt.read, ratio, tt.expected)
		}
	}
}

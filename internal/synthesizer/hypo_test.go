package synthesizer

import (
	"testing"
)

func TestEstimateIndexSize(t *testing.T) {
	bytes, pretty, penalty := EstimateIndexSize([]string{"user_id", "created_at"}, 1000000)
	if bytes <= 0 {
		t.Errorf("expected positive index size, got %d", bytes)
	}
	if pretty == "" {
		t.Error("expected non-empty pretty size string")
	}
	if penalty != "MODERATE" {
		t.Errorf("expected MODERATE penalty for 2 columns / 1M rows, got %s", penalty)
	}
}

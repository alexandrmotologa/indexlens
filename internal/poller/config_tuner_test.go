package poller

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

func TestCalculateTuning(t *testing.T) {
	profile := model.TuningProfile{
		TotalRAMGB:   32,
		VCPUs:        8,
		DiskType:     "NVMe",
		WorkloadType: "OLTP",
	}

	current := map[string]string{
		"shared_buffers":   "128MB",
		"random_page_cost": "4",
	}

	report := CalculateTuning(profile, current)

	if len(report.Settings) == 0 {
		t.Fatalf("expected non-empty settings list")
	}

	// 32GB RAM * 0.25 = 8192MB shared_buffers
	foundShared := false
	for _, s := range report.Settings {
		if s.Name == "shared_buffers" {
			foundShared = true
			if s.RecommendedValue != "8192MB" {
				t.Errorf("expected 8192MB shared_buffers, got %s", s.RecommendedValue)
			}
			if s.Status != "ATTENTION" {
				t.Errorf("expected ATTENTION status when current is 128MB, got %s", s.Status)
			}
		}
		if s.Name == "random_page_cost" {
			if s.RecommendedValue != "1.1" {
				t.Errorf("expected 1.1 for NVMe random_page_cost, got %s", s.RecommendedValue)
			}
		}
	}

	if !foundShared {
		t.Errorf("shared_buffers not found in report settings")
	}

	if !strings.Contains(report.AlterScript, "ALTER SYSTEM SET shared_buffers = '8192MB';") {
		t.Errorf("alter script missing expected ALTER SYSTEM statement")
	}
}

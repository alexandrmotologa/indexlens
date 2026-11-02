package poller

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

func TestConnectionPoolHealthStructure(t *testing.T) {
	health := model.ConnectionPoolHealth{
		TotalConnections:    142,
		MaxConnections:      200,
		UtilizationPercent:  71.0,
		ActiveCount:         38,
		IdleCount:           94,
		IdleInTxCount:       10,
		MaxIdleInTxDuration: 840.0,
		StateBreakdown: map[string]int{
			"active":              38,
			"idle":                94,
			"idle in transaction": 10,
		},
		OptimalPoolSize: 17,
		Status:          "WARNING",
		Warnings: []string{
			"Moderate connection saturation (71.0%).",
			"Detected 10 connection(s) 'idle in transaction' for up to 840 seconds.",
		},
		PgBouncerAdvice: "Current client connection count (142) significantly exceeds hardware sweet spot (17). Deploy PgBouncer in transaction pooling mode.",
	}

	if health.UtilizationPercent != 71.0 {
		t.Errorf("expected utilization 71.0, got %f", health.UtilizationPercent)
	}

	if !strings.Contains(health.PgBouncerAdvice, "PgBouncer") {
		t.Errorf("expected advice to mention PgBouncer")
	}
}

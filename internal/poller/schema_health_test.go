package poller

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

func TestSchemaHealthReportStructures(t *testing.T) {
	fk := model.UnindexedForeignKey{
		ConstraintName: "fk_orders_user_id",
		ChildTable:     "orders",
		ParentTable:    "users",
		Columns:        []string{"user_id"},
		SuggestedDDL:   "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fk_orders_user_id ON orders (user_id);",
		Severity:       "CRITICAL",
	}

	if !strings.Contains(fk.SuggestedDDL, "CONCURRENTLY") {
		t.Errorf("suggested DDL should use CONCURRENTLY")
	}

	report := model.SchemaHealthReport{
		UnindexedFKs: []model.UnindexedForeignKey{fk},
		MissingPKs: []model.TableMissingPK{
			{
				SchemaName:    "public",
				TableName:     "event_logs",
				EstimatedRows: 1500000,
			},
		},
		TotalIssues: 2,
		HealthScore: 75,
	}

	if report.HealthScore != 75 {
		t.Errorf("expected health score 75, got %d", report.HealthScore)
	}
}

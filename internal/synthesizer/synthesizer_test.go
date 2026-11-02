package synthesizer

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

func TestRankColumns(t *testing.T) {
	candidates := []ColumnCandidate{
		{ColumnName: "created_at", Type: PredRange},
		{ColumnName: "id", Type: PredSort, Direction: "DESC"},
		{ColumnName: "tenant_id", Type: PredEquality},
		{ColumnName: "status", Type: PredEquality},
	}

	ranked := RankColumns(candidates)
	if len(ranked) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(ranked))
	}

	// Equality first: tenant_id, status
	if ranked[0] != "tenant_id" || ranked[1] != "status" {
		t.Errorf("expected equality columns first, got %v", ranked[:2])
	}

	// Range second: created_at
	if ranked[2] != "created_at" {
		t.Errorf("expected range column third, got %s", ranked[2])
	}

	// Sort last: id DESC
	if ranked[3] != "id DESC" {
		t.Errorf("expected sort column fourth, got %s", ranked[3])
	}
}

func TestGenerateIndexName_Truncation(t *testing.T) {
	tableName := "very_long_transaction_history_table_name"
	columns := []string{"tenant_identifier_uuid", "created_timestamp_utc", "customer_reference_id", "status"}

	name := GenerateIndexName(tableName, columns)
	if len(name) > 63 {
		t.Errorf("generated index name exceeds PostgreSQL 63-byte limit: length=%d, name=%s", len(name), name)
	}

	if !strings.HasPrefix(name, "idx_") {
		t.Errorf("expected prefix idx_, got %s", name)
	}
}

func TestGenerateDDL(t *testing.T) {
	ddl := GenerateDDL("public", "orders", "idx_orders_tenant_created", []string{"tenant_id", "created_at DESC"}, "")
	expected := "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_tenant_created ON public.orders (tenant_id, created_at DESC);"
	if ddl != expected {
		t.Errorf("GenerateDDL mismatch:\nwant: %s\ngot:  %s", expected, ddl)
	}

	// With partial index
	ddlPartial := GenerateDDL("public", "users", "idx_users_email", []string{"email"}, "deleted_at IS NULL")
	expectedPartial := "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON public.users (email) WHERE deleted_at IS NULL;"
	if ddlPartial != expectedPartial {
		t.Errorf("GenerateDDL partial mismatch:\nwant: %s\ngot:  %s", expectedPartial, ddlPartial)
	}
}

func TestAdviseIndexes(t *testing.T) {
	tables := []model.TableStat{
		{
			SchemaName: "public",
			TableName:  "orders",
			SeqScan:    5000,
			SeqTupRead: 1500000,
			IdxScan:    200,
		},
	}

	queries := []model.QueryStat{
		{
			Query:        "SELECT * FROM orders WHERE tenant_id = $1 AND created_at >= $2 ORDER BY id DESC LIMIT 20",
			MeanExecTime: 85.0,
		},
	}

	recs := AdviseIndexes(tables, queries)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}

	r := recs[0]
	if r.TableName != "orders" {
		t.Errorf("expected recommendation for orders, got %s", r.TableName)
	}

	if !strings.Contains(r.DDL, "CONCURRENTLY") {
		t.Errorf("expected CONCURRENTLY in DDL, got %s", r.DDL)
	}

	if r.Priority != "HIGH" {
		t.Errorf("expected HIGH priority, got %s", r.Priority)
	}
}

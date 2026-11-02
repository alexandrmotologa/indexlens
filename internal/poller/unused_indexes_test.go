package poller

import (
	"testing"
)

func TestParseIndexColumns(t *testing.T) {
	def := "CREATE INDEX idx_orders_tenant ON public.orders USING btree (tenant_id, created_at DESC)"
	cols := parseIndexColumns(def)

	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0] != "tenant_id" {
		t.Errorf("expected tenant_id, got %s", cols[0])
	}
	if cols[1] != "created_at DESC" {
		t.Errorf("expected created_at DESC, got %s", cols[1])
	}
}

func TestIsPrefixIndex(t *testing.T) {
	shortCols := []string{"tenant_id"}
	longCols := []string{"tenant_id", "created_at DESC"}

	if !isPrefixIndex(shortCols, longCols) {
		t.Error("expected shortCols to be prefix of longCols")
	}

	differentCols := []string{"status"}
	if isPrefixIndex(differentCols, longCols) {
		t.Error("differentCols should not be prefix of longCols")
	}
}

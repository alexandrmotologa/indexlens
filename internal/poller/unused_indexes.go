package poller

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetUnusedIndexes detects zero-scan and redundant duplicate indexes.
func (c *Client) GetUnusedIndexes(ctx context.Context) ([]model.UnusedIndex, error) {
	query := `
		SELECT
			s.schemaname,
			s.relname AS table_name,
			s.indexrelname AS index_name,
			pg_relation_size(s.indexrelid) AS index_bytes,
			s.idx_scan,
			s.idx_tup_read,
			s.idx_tup_fetch,
			i.indisunique,
			i.indisprimary,
			pg_get_indexdef(s.indexrelid) AS index_def
		FROM pg_stat_user_indexes s
		JOIN pg_index i ON s.indexrelid = i.indexrelid
		JOIN pg_stat_user_tables t ON s.relid = t.relid
		WHERE NOT i.indisprimary
		  AND NOT i.indisunique
		  AND t.n_live_tup > 500
		ORDER BY s.idx_scan ASC, pg_relation_size(s.indexrelid) DESC
		LIMIT 100
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unused indexes: %w", err)
	}
	defer rows.Close()

	type rawIndex struct {
		schema     string
		table      string
		indexName  string
		bytes      int64
		scans      int64
		tupRead    int64
		tupFetch   int64
		isUnique   bool
		isPrimary  bool
		indexDef   string
		cols       []string
	}

	var rawIndexes []rawIndex
	tableIndexMap := make(map[string][]rawIndex)

	for rows.Next() {
		var r rawIndex
		if err := rows.Scan(
			&r.schema,
			&r.table,
			&r.indexName,
			&r.bytes,
			&r.scans,
			&r.tupRead,
			&r.tupFetch,
			&r.isUnique,
			&r.isPrimary,
			&r.indexDef,
		); err != nil {
			return nil, fmt.Errorf("scan error in unused indexes: %w", err)
		}

		r.cols = parseIndexColumns(r.indexDef)
		rawIndexes = append(rawIndexes, r)
		tableIndexMap[r.table] = append(tableIndexMap[r.table], r)
	}

	var unusedList []model.UnusedIndex

	for _, r := range rawIndexes {
		isRedundant := false
		redundantWith := ""

		// Check if another index on the same table has these columns as a leading prefix
		siblings := tableIndexMap[r.table]
		for _, sib := range siblings {
			if sib.indexName == r.indexName {
				continue
			}
			if isPrefixIndex(r.cols, sib.cols) {
				isRedundant = true
				redundantWith = fmt.Sprintf("%s (%s)", sib.indexName, strings.Join(sib.cols, ", "))
				break
			}
		}

		// Flag if 0 scans OR redundant
		if r.scans == 0 || isRedundant {
			dropDDL := fmt.Sprintf("DROP INDEX CONCURRENTLY IF EXISTS %s.%s;", r.schema, r.indexName)
			priority := "MEDIUM"
			if r.bytes > 50*1024*1024 || isRedundant {
				priority = "HIGH"
			}

			reason := fmt.Sprintf("%d scans recorded. Consumes %s disk space.", r.scans, model.FormatBytes(r.bytes))
			if isRedundant {
				reason = fmt.Sprintf("Redundant index. Leading prefix of %s.", redundantWith)
			}

			unusedList = append(unusedList, model.UnusedIndex{
				SchemaName:      r.schema,
				TableName:       r.table,
				IndexName:       r.indexName,
				IndexSizeBytes:  r.bytes,
				IndexSizePretty: model.FormatBytes(r.bytes),
				IndexScans:      r.scans,
				TuplesRead:      r.tupRead,
				TuplesFetched:   r.tupFetch,
				IsRedundant:     isRedundant,
				RedundantWith:   redundantWith,
				Columns:         r.cols,
				DropDDL:         dropDDL,
				Reason:          reason,
				Priority:        priority,
			})
		}
	}

	return unusedList, nil
}

func parseIndexColumns(indexDef string) []string {
	// Example: CREATE INDEX idx ON public.orders USING btree (tenant_id, created_at DESC)
	start := strings.Index(indexDef, "(")
	end := strings.LastIndex(indexDef, ")")
	if start == -1 || end == -1 || end <= start {
		return []string{}
	}

	colStr := indexDef[start+1 : end]
	parts := strings.Split(colStr, ",")
	var cols []string
	for _, p := range parts {
		clean := strings.TrimSpace(p)
		if clean != "" {
			cols = append(cols, clean)
		}
	}
	return cols
}

func isPrefixIndex(shortCols, longCols []string) bool {
	if len(shortCols) == 0 || len(shortCols) >= len(longCols) {
		return false
	}
	for i := range shortCols {
		// Clean out ASC/DESC
		s := strings.Fields(strings.ToLower(shortCols[i]))[0]
		l := strings.Fields(strings.ToLower(longCols[i]))[0]
		if s != l {
			return false
		}
	}
	return true
}

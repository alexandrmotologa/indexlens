package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetTableStats gathers sequential scans, index scans, dead tuples, and byte sizes.
func (c *Client) GetTableStats(ctx context.Context) ([]model.TableStat, error) {
	query := `
		SELECT
			schemaname,
			relname,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_live_tup,
			n_dead_tup,
			last_vacuum,
			last_autovacuum,
			pg_relation_size(relid) AS table_bytes,
			pg_indexes_size(relid) AS index_bytes
		FROM pg_stat_user_tables
		ORDER BY (seq_scan * 1000 + seq_tup_read) DESC
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_user_tables: %w", err)
	}
	defer rows.Close()

	var tables []model.TableStat
	for rows.Next() {
		var t model.TableStat
		var lastVac, lastAutoVac *time.Time

		err := rows.Scan(
			&t.SchemaName,
			&t.TableName,
			&t.SeqScan,
			&t.SeqTupRead,
			&t.IdxScan,
			&t.IdxTupFetch,
			&t.LiveTuples,
			&t.DeadTuples,
			&lastVac,
			&lastAutoVac,
			&t.TableSizeBytes,
			&t.IndexSizeBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error in table stats: %w", err)
		}

		t.LastVacuum = lastVac
		t.LastAutovacuum = lastAutoVac
		t.TableSizePretty = model.FormatBytes(t.TableSizeBytes)
		t.IndexSizePretty = model.FormatBytes(t.IndexSizeBytes)

		totalTuples := t.LiveTuples + t.DeadTuples
		if totalTuples > 0 {
			t.DeadTupleRatio = (float64(t.DeadTuples) / float64(totalTuples)) * 100.0
		}

		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error in table stats: %w", err)
	}

	return tables, nil
}

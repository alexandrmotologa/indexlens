package poller

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetCacheStats computes buffer cache hit ratios for tables and indexes.
func (c *Client) GetCacheStats(ctx context.Context) (model.CacheHitStats, error) {
	// 1. Global heap blocks
	queryHeap := `
		SELECT
			COALESCE(SUM(heap_blks_hit), 0) AS total_hit,
			COALESCE(SUM(heap_blks_read), 0) AS total_read
		FROM pg_statio_user_tables
	`
	var heapHit, heapRead int64
	if err := c.pool.QueryRow(ctx, queryHeap).Scan(&heapHit, &heapRead); err != nil {
		return model.CacheHitStats{}, fmt.Errorf("failed to query heap cache stats: %w", err)
	}

	// 2. Global index blocks
	queryIndex := `
		SELECT
			COALESCE(SUM(idx_blks_hit), 0) AS total_hit,
			COALESCE(SUM(idx_blks_read), 0) AS total_read
		FROM pg_statio_user_indexes
	`
	var idxHit, idxRead int64
	if err := c.pool.QueryRow(ctx, queryIndex).Scan(&idxHit, &idxRead); err != nil {
		return model.CacheHitStats{}, fmt.Errorf("failed to query index cache stats: %w", err)
	}

	// 3. Per-table cache stats for tables below 99% hit ratio
	queryTables := `
		SELECT
			t.relname AS table_name,
			t.heap_blks_hit,
			t.heap_blks_read,
			CASE WHEN (t.heap_blks_hit + t.heap_blks_read) > 0
				THEN (t.heap_blks_hit::float / (t.heap_blks_hit + t.heap_blks_read) * 100.0)
				ELSE 100.0 END AS heap_ratio,
			COALESCE(SUM(i.idx_blks_hit), 0) AS idx_hit,
			COALESCE(SUM(i.idx_blks_read), 0) AS idx_read,
			CASE WHEN (COALESCE(SUM(i.idx_blks_hit), 0) + COALESCE(SUM(i.idx_blks_read), 0)) > 0
				THEN (COALESCE(SUM(i.idx_blks_hit), 0)::float / (COALESCE(SUM(i.idx_blks_hit), 0) + COALESCE(SUM(i.idx_blks_read), 0)) * 100.0)
				ELSE 100.0 END AS idx_ratio
		FROM pg_statio_user_tables t
		LEFT JOIN pg_statio_user_indexes i ON t.relid = i.relid
		GROUP BY t.relname, t.heap_blks_hit, t.heap_blks_read
		HAVING (t.heap_blks_hit + t.heap_blks_read) > 500
		   AND (t.heap_blks_hit::float / NULLIF(t.heap_blks_hit + t.heap_blks_read, 0) * 100.0) < 99.0
		ORDER BY t.heap_blks_read DESC
		LIMIT 20
	`

	rows, err := c.pool.Query(ctx, queryTables)
	if err != nil {
		return model.CacheHitStats{}, fmt.Errorf("failed to query table cache stats: %w", err)
	}
	defer rows.Close()

	var tablesBelow []model.TableCacheStat
	for rows.Next() {
		var s model.TableCacheStat
		if err := rows.Scan(
			&s.TableName,
			&s.HeapBlksHit,
			&s.HeapBlksRead,
			&s.HeapHitRatio,
			&s.IndexBlksHit,
			&s.IndexBlksRead,
			&s.IndexHitRatio,
		); err != nil {
			return model.CacheHitStats{}, fmt.Errorf("scan error in table cache stats: %w", err)
		}
		tablesBelow = append(tablesBelow, s)
	}

	// Calculate overall and global ratios
	stats := model.CacheHitStats{
		TotalHeapHit:         heapHit,
		TotalHeapRead:        heapRead,
		TotalIndexHit:        idxHit,
		TotalIndexRead:       idxRead,
		TablesBelowThreshold: tablesBelow,
	}

	totalHeap := heapHit + heapRead
	if totalHeap > 0 {
		stats.GlobalHeapHitRatio = (float64(heapHit) / float64(totalHeap)) * 100.0
	} else {
		stats.GlobalHeapHitRatio = 100.0
	}

	totalIdx := idxHit + idxRead
	if totalIdx > 0 {
		stats.GlobalIndexHitRatio = (float64(idxHit) / float64(totalIdx)) * 100.0
	} else {
		stats.GlobalIndexHitRatio = 100.0
	}

	totalAll := totalHeap + totalIdx
	if totalAll > 0 {
		stats.OverallHitRatio = (float64(heapHit+idxHit) / float64(totalAll)) * 100.0
	} else {
		stats.OverallHitRatio = 100.0
	}

	return stats, nil
}

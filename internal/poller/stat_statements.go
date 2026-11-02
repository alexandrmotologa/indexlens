package poller

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/alexandrmotologa/indexlens/internal/explain"
	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetSlowQueries retrieves aggregated query performance metrics from pg_stat_statements.
func (c *Client) GetSlowQueries(ctx context.Context, limit int, sortBy string) ([]model.QueryStat, error) {
	if limit <= 0 {
		limit = 50
	}

	orderClause := "total_exec_time DESC"
	switch sortBy {
	case "mean":
		orderClause = "mean_exec_time DESC"
	case "calls":
		orderClause = "calls DESC"
	case "rows":
		orderClause = "rows DESC"
	case "cache_miss":
		orderClause = "shared_blks_read DESC"
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(queryid::text, '0') AS queryid,
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			min_exec_time,
			max_exec_time,
			stddev_exec_time,
			rows,
			shared_blks_hit,
			shared_blks_read,
			shared_blks_dirtied,
			shared_blks_written,
			local_blks_hit,
			local_blks_read,
			temp_blks_read,
			temp_blks_written
		FROM pg_stat_statements
		WHERE query NOT LIKE '%%pg_stat_statements%%'
		  AND query NOT LIKE '%%pg_stat_%%'
		ORDER BY %s
		LIMIT $1
	`, orderClause)

	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	var results []model.QueryStat
	for rows.Next() {
		var q model.QueryStat
		var qid string

		err := rows.Scan(
			&qid,
			&q.Query,
			&q.Calls,
			&q.TotalExecTime,
			&q.MeanExecTime,
			&q.MinExecTime,
			&q.MaxExecTime,
			&q.StddevExecTime,
			&q.Rows,
			&q.SharedBlksHit,
			&q.SharedBlksRead,
			&q.SharedBlksDirtied,
			&q.SharedBlksWritten,
			&q.LocalBlksHit,
			&q.LocalBlksRead,
			&q.TempBlksRead,
			&q.TempBlksWritten,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error in pg_stat_statements: %w", err)
		}

		q.QueryID = qid
		EnrichQueryMetrics(&q)

		results = append(results, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// EnrichQueryMetrics calculates cache hit ratio, jitter score, estimated latency percentiles,
// and identifies SQL anti-patterns.
func EnrichQueryMetrics(q *model.QueryStat) {
	if q.Calls > 0 {
		q.RowsPerCall = float64(q.Rows) / float64(q.Calls)
	}

	totalBlks := q.SharedBlksHit + q.SharedBlksRead
	if totalBlks > 0 {
		q.CacheHitRatio = (float64(q.SharedBlksHit) / float64(totalBlks)) * 100.0
	} else {
		q.CacheHitRatio = 100.0
	}

	// Calculate latency distribution and jitter
	if q.MeanExecTime > 0 {
		q.JitterScore = math.Round((q.StddevExecTime/q.MeanExecTime)*1000) / 1000
	}
	if q.JitterScore > 0.75 {
		q.JitterLevel = "SPIKY"
	} else if q.JitterScore > 0.25 {
		q.JitterLevel = "MODERATE"
	} else {
		q.JitterLevel = "STABLE"
	}

	// Percentile estimations based on mean and standard deviation
	p50 := q.MeanExecTime
	if q.StddevExecTime > 0 && q.MeanExecTime > q.MinExecTime {
		p50 = math.Max(q.MinExecTime, q.MeanExecTime-(0.2*q.StddevExecTime))
	}
	p90 := math.Min(q.MaxExecTime, q.MeanExecTime+(1.28*q.StddevExecTime))
	p95 := math.Min(q.MaxExecTime, q.MeanExecTime+(1.645*q.StddevExecTime))
	p99 := math.Min(q.MaxExecTime, q.MeanExecTime+(2.326*q.StddevExecTime))

	q.P50ExecTime = math.Round(p50*10) / 10
	q.P90ExecTime = math.Round(p90*10) / 10
	q.P95ExecTime = math.Round(p95*10) / 10
	q.P99ExecTime = math.Round(p99*10) / 10

	// Run anti-pattern inspection on query text
	q.AntiPatterns = explain.AnalyzeSQLAntiPatterns(q.Query)
}

// ResetStatStatements resets the statistics counters in pg_stat_statements.
func (c *Client) ResetStatStatements(ctx context.Context) error {
	_, err := c.pool.Exec(ctx, "SELECT pg_stat_statements_reset()")
	if err != nil {
		return fmt.Errorf("failed to reset pg_stat_statements: %w", err)
	}
	return nil
}

// ParseQueryID safely converts string query ID to int64.
func ParseQueryID(qid string) int64 {
	val, _ := strconv.ParseInt(qid, 10, 64)
	return val
}

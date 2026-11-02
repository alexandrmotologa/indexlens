package poller

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetConnectionPoolHealth inspects pg_stat_activity to detect connection saturation,
// idle-in-transaction leaks, and calculates optimal PgBouncer pool sizing.
func (c *Client) GetConnectionPoolHealth(ctx context.Context, vcpus int) (model.ConnectionPoolHealth, error) {
	if vcpus <= 0 {
		vcpus = 4
	}

	health := model.ConnectionPoolHealth{
		StateBreakdown: make(map[string]int),
		Warnings:       make([]string, 0),
		OptimalPoolSize: (vcpus * 2) + 1, // standard PgBouncer heuristic
	}

	// 1. Get max_connections
	var maxConnsStr string
	if err := c.pool.QueryRow(ctx, "SHOW max_connections").Scan(&maxConnsStr); err == nil {
		if val, err := strconv.Atoi(strings.TrimSpace(maxConnsStr)); err == nil {
			health.MaxConnections = val
		}
	}
	if health.MaxConnections <= 0 {
		health.MaxConnections = 100 // fallback
	}

	// 2. Query pg_stat_activity breakdown and idle in transaction durations
	query := `
		SELECT
			COALESCE(state, 'unknown') AS conn_state,
			count(*) AS state_count,
			COALESCE(MAX(EXTRACT(EPOCH FROM (clock_timestamp() - state_change))), 0) AS max_duration_sec
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'
		GROUP BY state
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return health, fmt.Errorf("failed to inspect pg_stat_activity: %w", err)
	}
	defer rows.Close()

	var totalConns int
	var maxIdleInTxSec float64

	for rows.Next() {
		var stateName string
		var count int
		var maxDur float64

		if err := rows.Scan(&stateName, &count, &maxDur); err == nil {
			health.StateBreakdown[stateName] = count
			totalConns += count

			switch strings.ToLower(stateName) {
			case "active":
				health.ActiveCount += count
			case "idle":
				health.IdleCount += count
			case "idle in transaction", "idle in transaction (aborted)":
				health.IdleInTxCount += count
				if maxDur > maxIdleInTxSec {
					maxIdleInTxSec = maxDur
				}
			}
		}
	}

	health.TotalConnections = totalConns
	health.MaxIdleInTxDuration = math.Round(maxIdleInTxSec*10) / 10
	if health.MaxConnections > 0 {
		health.UtilizationPercent = math.Round((float64(totalConns)/float64(health.MaxConnections))*1000) / 10
	}

	// Determine status and warnings
	health.Status = "HEALTHY"

	if health.UtilizationPercent >= 85.0 {
		health.Status = "CRITICAL"
		health.Warnings = append(health.Warnings, fmt.Sprintf("High connection saturation (%.1f%% of max_connections). Incoming clients risk receiving 'FATAL: sorry, too many clients already'.", health.UtilizationPercent))
	} else if health.UtilizationPercent >= 70.0 {
		health.Status = "WARNING"
		health.Warnings = append(health.Warnings, fmt.Sprintf("Moderate connection saturation (%.1f%%).", health.UtilizationPercent))
	}

	if health.IdleInTxCount > 0 && health.MaxIdleInTxDuration > 60.0 {
		health.Status = "WARNING"
		health.Warnings = append(health.Warnings, fmt.Sprintf("Detected %d connection(s) 'idle in transaction' for up to %.0f seconds. Long-lived transactions block vacuum tuple cleanup and risk table bloat.", health.IdleInTxCount, health.MaxIdleInTxDuration))
	}

	// PgBouncer sizing recommendation
	if totalConns > health.OptimalPoolSize*2 {
		health.PgBouncerAdvice = fmt.Sprintf("Current client connection count (%d) significantly exceeds hardware sweet spot (%d). Deploy PgBouncer in transaction pooling mode to cap server backends at %d and slash CPU context switching.", totalConns, health.OptimalPoolSize, health.OptimalPoolSize)
	} else {
		health.PgBouncerAdvice = fmt.Sprintf("Optimal direct PostgreSQL server connection limit is %d for a %d vCPU architecture. Connection load is currently within reasonable limits.", health.OptimalPoolSize, vcpus)
	}

	return health, nil
}

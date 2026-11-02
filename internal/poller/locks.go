package poller

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetLockContention retrieves running processes and builds the blocking dependency tree.
func (c *Client) GetLockContention(ctx context.Context) (model.LockContentionTree, error) {
	// Query active and idle in transaction connections along with blocking PIDs
	query := `
		SELECT
			a.pid,
			a.usename,
			a.datname,
			a.state,
			EXTRACT(EPOCH FROM (clock_timestamp() - a.query_start)) * 1000.0 AS duration_ms,
			COALESCE(a.wait_event_type, '') AS wait_event_type,
			COALESCE(a.wait_event, '') AS wait_event,
			a.query,
			COALESCE((pg_blocking_pids(a.pid))[1], 0) AS blocking_pid
		FROM pg_stat_activity a
		WHERE a.pid != pg_backend_pid()
		  AND a.state != 'idle'
		ORDER BY duration_ms DESC
		LIMIT 100
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return model.LockContentionTree{}, fmt.Errorf("failed to query pg_stat_activity for locks: %w", err)
	}
	defer rows.Close()

	var allProcs []model.LiveActivityItem
	blockerMap := make(map[int64][]int64)
	isBlockedMap := make(map[int64]bool)

	for rows.Next() {
		var item model.LiveActivityItem
		var blockingPID int64

		if err := rows.Scan(
			&item.PID,
			&item.User,
			&item.Database,
			&item.State,
			&item.DurationMs,
			&item.WaitEventType,
			&item.WaitEvent,
			&item.Query,
			&blockingPID,
		); err != nil {
			return model.LockContentionTree{}, fmt.Errorf("scan error in lock activity: %w", err)
		}

		item.BlockingPID = blockingPID
		if blockingPID > 0 {
			blockerMap[blockingPID] = append(blockerMap[blockingPID], item.PID)
			isBlockedMap[item.PID] = true
		}

		allProcs = append(allProcs, item)
	}

	// Update items with their blocked PIDs
	var rootBlockers []model.LiveActivityItem
	totalBlocked := 0

	for i := range allProcs {
		p := &allProcs[i]
		if children, exists := blockerMap[p.PID]; exists {
			p.BlockedPIDs = children
			totalBlocked += len(children)
			if !isBlockedMap[p.PID] {
				rootBlockers = append(rootBlockers, *p)
			}
		}
	}

	// Count active locks in pg_locks
	var activeLocksCount int
	_ = c.pool.QueryRow(ctx, "SELECT count(*) FROM pg_locks WHERE NOT granted OR mode LIKE '%Exclusive%'").Scan(&activeLocksCount)

	return model.LockContentionTree{
		RootBlockers:        rootBlockers,
		ActiveProcesses:     allProcs,
		TotalBlockedQueries: totalBlocked,
		ActiveLocksCount:    activeLocksCount,
		HasContention:       len(rootBlockers) > 0,
	}, nil
}

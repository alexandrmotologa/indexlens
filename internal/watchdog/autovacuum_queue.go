package watchdog

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
)

// GetAutovacuumWorkers queries pg_stat_progress_vacuum for running vacuum tasks.
func GetAutovacuumWorkers(ctx context.Context, client *poller.Client) ([]model.AutovacuumWorker, error) {
	query := `
		SELECT
			p.pid,
			d.datname,
			c.relname,
			p.phase,
			p.heap_blks_total,
			p.heap_blks_scanned,
			p.heap_blks_vacuumed,
			p.index_vacuum_count
		FROM pg_stat_progress_vacuum p
		JOIN pg_database d ON p.datid = d.oid
		JOIN pg_class c ON p.relid = c.oid
		ORDER BY p.pid ASC
	`

	rows, err := client.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_progress_vacuum: %w", err)
	}
	defer rows.Close()

	var workers []model.AutovacuumWorker
	for rows.Next() {
		var w model.AutovacuumWorker
		err := rows.Scan(
			&w.PID,
			&w.Database,
			&w.Table,
			&w.Phase,
			&w.HeapBlksTotal,
			&w.HeapBlksScanned,
			&w.HeapBlksVacuumed,
			&w.IndexVacuumCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error in vacuum progress: %w", err)
		}

		if w.HeapBlksTotal > 0 {
			w.ProgressPercentage = (float64(w.HeapBlksScanned) / float64(w.HeapBlksTotal)) * 100.0
		}

		workers = append(workers, w)
	}

	return workers, nil
}

package watchdog

import (
	"context"
	"fmt"
	"math"

	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
)

// EstimateTableBloat inspects live tables and estimates wasted page bloat from dead tuples.
func EstimateTableBloat(ctx context.Context, client *poller.Client) ([]model.BloatInfo, error) {
	query := `
		SELECT
			schemaname || '.' || relname AS full_table_name,
			pg_total_relation_size(relid) AS total_bytes,
			n_live_tup,
			n_dead_tup
		FROM pg_stat_user_tables
		WHERE (n_live_tup + n_dead_tup) > 100
		ORDER BY n_dead_tup DESC
		LIMIT 50
	`

	rows, err := client.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table bloat statistics: %w", err)
	}
	defer rows.Close()

	var bloatList []model.BloatInfo
	for rows.Next() {
		var b model.BloatInfo
		err := rows.Scan(
			&b.TableName,
			&b.SizeBytes,
			&b.LiveTuples,
			&b.DeadTuples,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error in bloat statistics: %w", err)
		}

		totalTuples := b.LiveTuples + b.DeadTuples
		if totalTuples > 0 {
			b.DeadTupleRatio = (float64(b.DeadTuples) / float64(totalTuples)) * 100.0
			// Estimate bloat bytes proportional to dead tuple ratio, damped by standard 10% fill factor
			dampedRatio := math.Max(0, b.DeadTupleRatio-10.0) / 100.0
			b.BloatBytes = int64(float64(b.SizeBytes) * dampedRatio)
			if b.SizeBytes > 0 {
				b.BloatRatio = (float64(b.BloatBytes) / float64(b.SizeBytes)) * 100.0
			}
		}

		b.SizePretty = model.FormatBytes(b.SizeBytes)
		b.BloatPretty = model.FormatBytes(b.BloatBytes)

		bloatList = append(bloatList, b)
	}

	return bloatList, nil
}

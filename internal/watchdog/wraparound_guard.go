package watchdog

import (
	"context"
	"fmt"

	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
)

// CheckWraparound calculates txid age relative to autovacuum_freeze_max_age.
func CheckWraparound(ctx context.Context, client *poller.Client) (model.WraparoundInfo, error) {
	// 1. Get autovacuum_freeze_max_age setting
	var maxAge int64 = 200000000 // default PostgreSQL setting
	var maxAgeStr string
	if err := client.Pool().QueryRow(ctx, "SHOW autovacuum_freeze_max_age").Scan(&maxAgeStr); err == nil {
		_, _ = fmt.Sscanf(maxAgeStr, "%d", &maxAge)
	}

	// 2. Query oldest transaction ID age
	query := `
		SELECT
			datname,
			age(datfrozenxid) AS txid_age
		FROM pg_database
		WHERE datallowconn = true
		ORDER BY age(datfrozenxid) DESC
		LIMIT 1
	`

	var dbName string
	var currentAge int64
	err := client.Pool().QueryRow(ctx, query).Scan(&dbName, &currentAge)
	if err != nil {
		return model.WraparoundInfo{}, fmt.Errorf("failed to query database txid age: %w", err)
	}

	percentConsumed := (float64(currentAge) / float64(maxAge)) * 100.0
	remaining := maxAge - currentAge
	if remaining < 0 {
		remaining = 0
	}

	status := "HEALTHY"
	if percentConsumed >= 75.0 || currentAge >= 150000000 {
		status = "CRITICAL"
	} else if percentConsumed >= 50.0 || currentAge >= 100000000 {
		status = "WARNING"
	}

	return model.WraparoundInfo{
		DatabaseName:    dbName,
		CurrentTxidAge:  currentAge,
		MaxTxidAge:      maxAge,
		PercentConsumed: percentConsumed,
		RemainingTxids:  remaining,
		Status:          status,
	}, nil
}

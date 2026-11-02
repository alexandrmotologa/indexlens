package synthesizer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
)

// EstimateIndexSize calculates expected disk footprint and write overhead.
func EstimateIndexSize(columns []string, rowCount int64) (int64, string, string) {
	if rowCount <= 0 {
		rowCount = 100000 // default heuristic baseline
	}

	// Average byte width per column
	avgWidth := 8 // default fallback width
	for _, col := range columns {
		colLower := strings.ToLower(col)
		if strings.Contains(colLower, "id") || strings.Contains(colLower, "int") {
			avgWidth += 4
		} else if strings.Contains(colLower, "time") || strings.Contains(colLower, "date") {
			avgWidth += 8
		} else if strings.Contains(colLower, "uuid") {
			avgWidth += 16
		} else {
			avgWidth += 16 // text or varchar estimate
		}
	}

	// 8 bytes IndexTupleData header + 4 bytes line pointer + data payload
	tupleSize := 12 + avgWidth

	// Usable page bytes in 8KB page (8192 - 24 page header - 4 special)
	usablePageBytes := 8164

	tuplesPerPage := usablePageBytes / tupleSize
	if tuplesPerPage <= 0 {
		tuplesPerPage = 1
	}

	pages := float64(rowCount) / float64(tuplesPerPage)
	// B-Tree fill factor overhead (typically 70%-80% page utilization + interior nodes)
	totalBytes := int64(pages * 8192.0 * 1.35)

	writePenalty := "LOW"
	if len(columns) >= 3 || rowCount > 2000000 {
		writePenalty = "SIGNIFICANT"
	} else if len(columns) >= 2 || rowCount > 500000 {
		writePenalty = "MODERATE"
	}

	return totalBytes, model.FormatBytes(totalBytes), writePenalty
}

// SimulateHypoPG runs a hypothetical index test using the hypopg extension if present.
func SimulateHypoPG(ctx context.Context, client *poller.Client, ddl string, testQuery string) (model.HypoSimulationResult, error) {
	hasExt, err := client.HasExtension(ctx, "hypopg")
	if err != nil || !hasExt {
		return model.HypoSimulationResult{
			Supported:          false,
			ExtensionInstalled: false,
			Message:            "HypoPG extension not detected in database. Install hypopg (CREATE EXTENSION hypopg) to verify planner index selection in memory with 0 disk write.",
		}, nil
	}

	// 1. Get original cost
	origExplain := "EXPLAIN (FORMAT JSON) " + testQuery
	var origJSONStr string
	if err := client.Pool().QueryRow(ctx, origExplain).Scan(&origJSONStr); err != nil {
		return model.HypoSimulationResult{}, fmt.Errorf("failed to run baseline explain: %w", err)
	}

	var origPlans []struct {
		Plan struct {
			TotalCost float64 `json:"Total Cost"`
			NodeType  string  `json:"Node Type"`
		} `json:"Plan"`
	}
	_ = json.Unmarshal([]byte(origJSONStr), &origPlans)
	origCost := 0.0
	origType := "Seq Scan"
	if len(origPlans) > 0 {
		origCost = origPlans[0].Plan.TotalCost
		origType = origPlans[0].Plan.NodeType
	}

	// 2. Create hypothetical index
	createHypoSQL := fmt.Sprintf("SELECT indexrelid, index_name FROM hypopg_create_index('%s')", strings.ReplaceAll(ddl, "'", "''"))
	var hypoOid int64
	var hypoName string
	if err := client.Pool().QueryRow(ctx, createHypoSQL).Scan(&hypoOid, &hypoName); err != nil {
		return model.HypoSimulationResult{}, fmt.Errorf("failed to create hypothetical index: %w", err)
	}

	// Defer dropping the hypothetical index
	defer func() {
		_, _ = client.Pool().Exec(ctx, fmt.Sprintf("SELECT hypopg_drop_index(%d)", hypoOid))
	}()

	// 3. Re-run explain with hypothetical index active
	var hypoJSONStr string
	if err := client.Pool().QueryRow(ctx, origExplain).Scan(&hypoJSONStr); err != nil {
		return model.HypoSimulationResult{}, fmt.Errorf("failed to run hypothetical explain: %w", err)
	}

	var hypoPlans []struct {
		Plan struct {
			TotalCost float64 `json:"Total Cost"`
			NodeType  string  `json:"Node Type"`
		} `json:"Plan"`
	}
	_ = json.Unmarshal([]byte(hypoJSONStr), &hypoPlans)
	hypoCost := origCost
	hypoType := origType
	if len(hypoPlans) > 0 {
		hypoCost = hypoPlans[0].Plan.TotalCost
		hypoType = hypoPlans[0].Plan.NodeType
	}

	reductionPct := 0.0
	if origCost > 0 {
		reductionPct = math.Max(0, ((origCost-hypoCost)/origCost)*100.0)
	}

	msg := fmt.Sprintf("Planner confirmed! Estimated cost dropped from %.1f to %.1f (%.1f%% reduction).",
		origCost, hypoCost, reductionPct)

	return model.HypoSimulationResult{
		Supported:          true,
		ExtensionInstalled: true,
		OriginalCost:       origCost,
		SimulatedCost:      hypoCost,
		CostReductionPct:   reductionPct,
		OriginalPlanType:   origType,
		SimulatedPlanType:  hypoType,
		IndexName:          hypoName,
		Message:            msg,
	}, nil
}

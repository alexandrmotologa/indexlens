package explain

import (
	"fmt"
	"math"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// DetectBottlenecks inspects all nodes in an execution plan and extracts performance red flags.
func DetectBottlenecks(root *model.ExplainPlanNode, nodes []*model.ExplainPlanNode, totalTime float64) []model.Bottleneck {
	var bottlenecks []model.Bottleneck

	for _, n := range nodes {
		// 1. Disk Spill in Sort or Hash nodes
		if strings.Contains(strings.ToLower(n.SortMethod), "disk") ||
			strings.EqualFold(n.SortSpaceType, "disk") ||
			n.TempWrittenBlocks > 0 {
			spillKB := n.SortSpaceUsed
			if spillKB == 0 && n.TempWrittenBlocks > 0 {
				spillKB = n.TempWrittenBlocks * 8 // 8KB per block in Postgres
			}
			msg := fmt.Sprintf("Node spilled %d KB to temporary disk storage (%s)", spillKB, n.SortMethod)
			if n.SortMethod == "" {
				msg = fmt.Sprintf("Node wrote %d temporary disk blocks (%d KB)", n.TempWrittenBlocks, spillKB)
			}
			bottlenecks = append(bottlenecks, model.Bottleneck{
				Type:           "DISK_SPILL",
				Severity:       "CRITICAL",
				NodeID:         n.NodeID,
				NodeType:       n.NodeType,
				Relation:       n.RelationName,
				Message:        msg,
				Recommendation: "Increase work_mem or add an index that supplies the required sort ordering to avoid memory spilling",
				ImpactScore:    80.0,
			})
		}

		// 2. Heavy Sequential Scan
		if strings.EqualFold(n.NodeType, "Seq Scan") {
			totalScanned := n.ActualRows + float64(n.RowsRemovedByFilter)
			if totalScanned > 500 || n.ActualTotalTime > 15.0 {
				rejectionPercent := 0.0
				if totalScanned > 0 {
					rejectionPercent = (float64(n.RowsRemovedByFilter) / totalScanned) * 100.0
				}

				severity := "WARNING"
				if totalScanned > 10000 || rejectionPercent > 80.0 {
					severity = "CRITICAL"
				}

				msg := fmt.Sprintf("Sequential scan on table '%s' read %.0f rows (took %.1f ms)", n.RelationName, totalScanned, n.ActualTotalTime)
				if n.RowsRemovedByFilter > 0 {
					msg = fmt.Sprintf("Sequential scan on '%s' examined %.0f rows and discarded %d rows (%.1f%% rejection rate)",
						n.RelationName, totalScanned, n.RowsRemovedByFilter, rejectionPercent)
				}

				bottlenecks = append(bottlenecks, model.Bottleneck{
					Type:           "HEAVY_SEQ_SCAN",
					Severity:       severity,
					NodeID:         n.NodeID,
					NodeType:       n.NodeType,
					Relation:       n.RelationName,
					Message:        msg,
					Recommendation: fmt.Sprintf("Create an index on table '%s' targeting filtered columns: %s", n.RelationName, cleanFilter(n.Filter)),
					ImpactScore:    math.Min(95.0, 50.0+(totalScanned/1000.0)),
				})
			}
		}

		// 3. Dominant Time Sink (> 40% of overall query time)
		if totalTime > 10.0 && n.PercentageOfTotal >= 40.0 {
			bottlenecks = append(bottlenecks, model.Bottleneck{
				Type:           "TIME_SINK",
				Severity:       "WARNING",
				NodeID:         n.NodeID,
				NodeType:       n.NodeType,
				Relation:       n.RelationName,
				Message:        fmt.Sprintf("Node accounts for %.1f%% (%.1f ms) of total query duration", n.PercentageOfTotal, n.ExclusiveTime),
				Recommendation: "Focus optimization efforts on this operation to achieve the highest latency reduction",
				ImpactScore:    n.PercentageOfTotal,
			})
		}

		// 4. Stale Planner Statistics (large row count discrepancy)
		if n.PlanRows > 0 && n.ActualRows > 50 {
			ratio := n.ActualRows / n.PlanRows
			if ratio > 10.0 || ratio < 0.1 {
				bottlenecks = append(bottlenecks, model.Bottleneck{
					Type:           "STALE_STATS",
					Severity:       "WARNING",
					NodeID:         n.NodeID,
					NodeType:       n.NodeType,
					Relation:       n.RelationName,
					Message:        fmt.Sprintf("Planner estimated %.0f rows, but execution returned %.0f rows (%.1fx discrepancy)", n.PlanRows, n.ActualRows, ratio),
					Recommendation: fmt.Sprintf("Run 'ANALYZE %s' to refresh planner distribution statistics and prevent sub-optimal join plans", n.RelationName),
					ImpactScore:    65.0,
				})
			}
		}

		// 5. Heavy Cache Miss (shared blocks read from storage)
		totalBlocks := n.SharedHitBlocks + n.SharedReadBlocks
		if totalBlocks > 100 {
			readRatio := float64(n.SharedReadBlocks) / float64(totalBlocks)
			if readRatio > 0.30 {
				bottlenecks = append(bottlenecks, model.Bottleneck{
					Type:           "CACHE_MISS",
					Severity:       "WARNING",
					NodeID:         n.NodeID,
					NodeType:       n.NodeType,
					Relation:       n.RelationName,
					Message:        fmt.Sprintf("High disk read ratio: %d of %d blocks (%.1f%%) fetched from storage", n.SharedReadBlocks, totalBlocks, readRatio*100),
					Recommendation: "Evaluate table working set size vs shared_buffers or add a selective index to reduce pages scanned",
					ImpactScore:    50.0 + (readRatio * 40.0),
				})
			}
		}
	}

	return bottlenecks
}

func cleanFilter(f string) string {
	if f == "" {
		return "matching query WHERE clause"
	}
	f = strings.ReplaceAll(f, "::timestamp without time zone", "")
	f = strings.ReplaceAll(f, "::text", "")
	f = strings.ReplaceAll(f, "::integer", "")
	return f
}

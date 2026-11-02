package explain

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// ComparePlans calculates structural, timing, and buffer differences between two plans.
func ComparePlans(baseline, current *model.ExplainAnalysisResult) model.PlanDiffResult {
	if baseline == nil || current == nil {
		return model.PlanDiffResult{
			Verdict: "NEUTRAL",
			Summary: "Insufficient plan data for comparison",
		}
	}

	timeDelta := current.TotalExecutionTime - baseline.TotalExecutionTime
	var timeDeltaPct float64
	if baseline.TotalExecutionTime > 0 {
		timeDeltaPct = (timeDelta / baseline.TotalExecutionTime) * 100.0
	}

	verdict := "NEUTRAL"
	if timeDeltaPct <= -15.0 {
		verdict = "IMPROVED"
	} else if timeDeltaPct >= 15.0 {
		verdict = "REGRESSED"
	}

	// Compare root / key nodes
	var nodeDiffs []model.PlanNodeDiff

	minLen := int(math.Min(float64(len(baseline.FlattenedNodes)), float64(len(current.FlattenedNodes))))
	for i := 0; i < minLen; i++ {
		bNode := baseline.FlattenedNodes[i]
		cNode := current.FlattenedNodes[i]

		tDelta := cNode.ActualTotalTime - bNode.ActualTotalTime
		if math.Abs(tDelta) > 1.0 || bNode.NodeType != cNode.NodeType {
			summary := fmt.Sprintf("Time changed by %.1f ms", tDelta)
			if bNode.NodeType != cNode.NodeType {
				summary = fmt.Sprintf("Node transformed from '%s' to '%s' (delta %.1f ms)", bNode.NodeType, cNode.NodeType, tDelta)
			}

			nodeDiffs = append(nodeDiffs, model.PlanNodeDiff{
				Path:          fmt.Sprintf("node-%d", i+1),
				BaselineType:  bNode.NodeType,
				CurrentType:   cNode.NodeType,
				BaselineTime:  bNode.ActualTotalTime,
				CurrentTime:   cNode.ActualTotalTime,
				TimeDeltaMs:   tDelta,
				ChangeSummary: summary,
			})
		}
	}

	summaryText := fmt.Sprintf("Execution duration shifted from %.1f ms to %.1f ms (%+.1f%% delta).",
		baseline.TotalExecutionTime, current.TotalExecutionTime, timeDeltaPct)

	if verdict == "IMPROVED" {
		summaryText = fmt.Sprintf("Execution performance improved by %.1f%% (%.1f ms -> %.1f ms). Disk reads reduced by %d blocks.",
			math.Abs(timeDeltaPct), baseline.TotalExecutionTime, current.TotalExecutionTime, baseline.TotalBuffersRead-current.TotalBuffersRead)
	} else if verdict == "REGRESSED" {
		summaryText = fmt.Sprintf("Execution regressed by +%.1f%% (%.1f ms -> %.1f ms). Increased disk reads or higher node cost detected.",
			timeDeltaPct, baseline.TotalExecutionTime, current.TotalExecutionTime)
	}

	return model.PlanDiffResult{
		BaselineExecutionTime: baseline.TotalExecutionTime,
		CurrentExecutionTime:  current.TotalExecutionTime,
		TimeDeltaMs:           timeDelta,
		TimeDeltaPct:          timeDeltaPct,
		BaselineBuffersHit:    baseline.TotalBuffersHit,
		CurrentBuffersHit:     current.TotalBuffersHit,
		BaselineBuffersRead:   baseline.TotalBuffersRead,
		CurrentBuffersRead:    current.TotalBuffersRead,
		NodeChanges:           nodeDiffs,
		Verdict:               verdict,
		Summary:               summaryText,
	}
}

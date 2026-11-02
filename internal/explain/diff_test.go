package explain

import (
	"testing"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

func TestComparePlans(t *testing.T) {
	baseline := &model.ExplainAnalysisResult{
		TotalExecutionTime: 182.6,
		TotalBuffersHit:    1000,
		TotalBuffersRead:   5000,
		FlattenedNodes: []*model.ExplainPlanNode{
			{NodeType: "Seq Scan", ActualTotalTime: 182.6},
		},
	}

	current := &model.ExplainAnalysisResult{
		TotalExecutionTime: 4.2,
		TotalBuffersHit:    800,
		TotalBuffersRead:   0,
		FlattenedNodes: []*model.ExplainPlanNode{
			{NodeType: "Index Scan", ActualTotalTime: 4.2},
		},
	}

	diff := ComparePlans(baseline, current)
	if diff.Verdict != "IMPROVED" {
		t.Errorf("expected verdict IMPROVED, got %s", diff.Verdict)
	}

	if diff.TimeDeltaMs >= 0 {
		t.Errorf("expected negative time delta, got %f", diff.TimeDeltaMs)
	}

	if len(diff.NodeChanges) != 1 {
		t.Fatalf("expected 1 node change, got %d", len(diff.NodeChanges))
	}
}

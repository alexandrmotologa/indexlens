package explain

import (
	"github.com/alexandrmotologa/indexlens/internal/model"
)

// FlamegraphBar represents a visual horizontal slice in an execution flamegraph.
type FlamegraphBar struct {
	NodeID          string  `json:"nodeId"`
	NodeType        string  `json:"nodeType"`
	RelationName    string  `json:"relationName,omitempty"`
	Depth           int     `json:"depth"`
	LeftOffsetPct   float64 `json:"leftOffsetPct"` // 0.0 to 100.0
	WidthPct        float64 `json:"widthPct"`      // 0.0 to 100.0
	ExclusiveTime   float64 `json:"exclusiveTime"` // ms
	ActualTotalTime float64 `json:"actualTotalTime"`
	SharedHitBlocks int64   `json:"sharedHitBlocks"`
	SharedReadBlocks int64  `json:"sharedReadBlocks"`
	HighlightColor  string  `json:"highlightColor"`
}

// BuildFlamegraphStructure constructs positioned bars for rendering visual flamegraphs.
func BuildFlamegraphStructure(root *model.ExplainPlanNode, totalTime float64) []FlamegraphBar {
	if root == nil || totalTime <= 0 {
		return nil
	}

	var bars []FlamegraphBar
	var walk func(node *model.ExplainPlanNode, depth int, leftOffsetPct, availableWidthPct float64)

	walk = func(node *model.ExplainPlanNode, depth int, leftOffsetPct, availableWidthPct float64) {
		widthPct := (node.ActualTotalTime / totalTime) * 100.0
		if widthPct > availableWidthPct {
			widthPct = availableWidthPct
		}
		if widthPct < 0.5 {
			widthPct = 0.5 // minimum visible width
		}

		color := "#3b82f6" // default blue
		if node.NodeType == "Seq Scan" {
			color = "#ef4444" // red for seq scan
		} else if node.NodeType == "Sort" && (node.SortSpaceType == "Disk" || node.TempWrittenBlocks > 0) {
			color = "#f97316" // orange for disk sort
		} else if node.PercentageOfTotal >= 40.0 {
			color = "#eab308" // yellow for time sink
		} else if node.NodeType == "Index Scan" || node.NodeType == "Index Only Scan" {
			color = "#10b981" // green for index scan
		}

		bars = append(bars, FlamegraphBar{
			NodeID:           node.NodeID,
			NodeType:         node.NodeType,
			RelationName:     node.RelationName,
			Depth:            depth,
			LeftOffsetPct:    leftOffsetPct,
			WidthPct:         widthPct,
			ExclusiveTime:    node.ExclusiveTime,
			ActualTotalTime:  node.ActualTotalTime,
			SharedHitBlocks:  node.SharedHitBlocks,
			SharedReadBlocks: node.SharedReadBlocks,
			HighlightColor:   color,
		})

		currentLeft := leftOffsetPct
		for _, child := range node.Plans {
			childWidth := (child.ActualTotalTime / totalTime) * 100.0
			walk(child, depth+1, currentLeft, widthPct)
			currentLeft += childWidth
		}
	}

	walk(root, 0, 0.0, 100.0)
	return bars
}

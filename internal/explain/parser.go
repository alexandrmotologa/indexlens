package explain

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// rawPlanWrapper represents the top-level structure of PostgreSQL EXPLAIN JSON output.
type rawPlanWrapper struct {
	Plan           rawPlanNode `json:"Plan"`
	PlanningTime   float64     `json:"Planning Time"`
	ExecutionTime  float64     `json:"Execution Time"`
	Triggers       []any       `json:"Triggers,omitempty"`
}

// rawPlanNode represents the raw JSON fields from PostgreSQL EXPLAIN.
type rawPlanNode struct {
	NodeType            string        `json:"Node Type"`
	RelationName        string        `json:"Relation Name,omitempty"`
	Schema              string        `json:"Schema,omitempty"`
	Alias               string        `json:"Alias,omitempty"`
	IndexName           string        `json:"Index Name,omitempty"`
	StartupCost         float64       `json:"Startup Cost"`
	TotalCost           float64       `json:"Total Cost"`
	PlanRows            float64       `json:"Plan Rows"`
	PlanWidth           int           `json:"Plan Width"`
	ActualStartupTime   float64       `json:"Actual Startup Time"`
	ActualTotalTime     float64       `json:"Actual Total Time"`
	ActualRows          float64       `json:"Actual Rows"`
	ActualLoops         float64       `json:"Actual Loops"`
	SharedHitBlocks     int64         `json:"Shared Hit Blocks"`
	SharedReadBlocks    int64         `json:"Shared Read Blocks"`
	SharedDirtiedBlocks int64         `json:"Shared Dirtied Blocks"`
	SharedWrittenBlocks int64         `json:"Shared Written Blocks"`
	TempReadBlocks      int64         `json:"Temp Read Blocks"`
	TempWrittenBlocks   int64         `json:"Temp Written Blocks"`
	SortMethod          string        `json:"Sort Method,omitempty"`
	SortSpaceUsed       int64         `json:"Sort Space Used,omitempty"`
	SortSpaceType       string        `json:"Sort Space Type,omitempty"`
	Filter              string        `json:"Filter,omitempty"`
	IndexCond           string        `json:"Index Cond,omitempty"`
	HashCond            string        `json:"Hash Cond,omitempty"`
	RecheckCond         string        `json:"Recheck Cond,omitempty"`
	RowsRemovedByFilter int64         `json:"Rows Removed by Filter,omitempty"`
	Plans               []rawPlanNode `json:"Plans,omitempty"`
}

// ParsePlanJSON parses raw JSON string into a structured ExplainAnalysisResult.
func ParsePlanJSON(jsonBytes []byte) (*model.ExplainAnalysisResult, error) {
	trimmed := strings.TrimSpace(string(jsonBytes))
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty plan input")
	}

	var wrapper rawPlanWrapper

	// Support both array-wrapped [{ Plan: ... }] and direct { Plan: ... } formats
	if strings.HasPrefix(trimmed, "[") {
		var wrappers []rawPlanWrapper
		if err := json.Unmarshal([]byte(trimmed), &wrappers); err != nil {
			return nil, fmt.Errorf("failed to parse explain JSON array: %w", err)
		}
		if len(wrappers) == 0 {
			return nil, fmt.Errorf("explain JSON array is empty")
		}
		wrapper = wrappers[0]
	} else {
		// Try direct wrapper or naked Plan object
		if err := json.Unmarshal([]byte(trimmed), &wrapper); err != nil {
			var rawNode rawPlanNode
			if errNode := json.Unmarshal([]byte(trimmed), &rawNode); errNode == nil && rawNode.NodeType != "" {
				wrapper = rawPlanWrapper{Plan: rawNode}
			} else {
				return nil, fmt.Errorf("failed to parse explain JSON object: %w", err)
			}
		}
	}

	if wrapper.Plan.NodeType == "" {
		return nil, fmt.Errorf("invalid plan: missing Node Type")
	}

	nodeCounter := 0
	root := convertNode(&wrapper.Plan, &nodeCounter)

	// Determine total execution time
	totalTime := wrapper.ExecutionTime
	if totalTime <= 0 {
		totalTime = root.ActualTotalTime
	}
	if totalTime <= 0 {
		totalTime = root.TotalCost // fallback if ANALYZE was omitted
	}

	// Compute exclusive times and percentage of total duration
	computeExclusiveTimes(root, totalTime)

	// Flatten nodes and aggregate buffer statistics
	var flatNodes []*model.ExplainPlanNode
	var totalHit, totalRead int64

	var collect func(n *model.ExplainPlanNode)
	collect = func(n *model.ExplainPlanNode) {
		flatNodes = append(flatNodes, n)
		totalHit += n.SharedHitBlocks
		totalRead += n.SharedReadBlocks
		for _, child := range n.Plans {
			collect(child)
		}
	}
	collect(root)

	var cacheHitRatio float64 = 100.0
	totalBufs := totalHit + totalRead
	if totalBufs > 0 {
		cacheHitRatio = (float64(totalHit) / float64(totalBufs)) * 100.0
	}

	// Detect bottlenecks across all nodes
	bottlenecks := DetectBottlenecks(root, flatNodes, totalTime)

	return &model.ExplainAnalysisResult{
		RootNode:           root,
		TotalExecutionTime: totalTime,
		PlanningTime:       wrapper.PlanningTime,
		TotalBuffersHit:    totalHit,
		TotalBuffersRead:   totalRead,
		CacheHitRatio:      cacheHitRatio,
		Bottlenecks:        bottlenecks,
		FlattenedNodes:     flatNodes,
	}, nil
}

func convertNode(raw *rawPlanNode, counter *int) *model.ExplainPlanNode {
	*counter++
	nodeID := fmt.Sprintf("node-%d", *counter)

	// Normalize loop multiplier if present
	loops := raw.ActualLoops
	if loops <= 0 {
		loops = 1
	}

	// In PostgreSQL, Actual Total Time is the average per loop; multiply to get overall time
	totalTime := raw.ActualTotalTime * loops

	node := &model.ExplainPlanNode{
		NodeID:              nodeID,
		NodeType:            raw.NodeType,
		RelationName:        raw.RelationName,
		Schema:              raw.Schema,
		Alias:               raw.Alias,
		IndexName:           raw.IndexName,
		StartupCost:         raw.StartupCost,
		TotalCost:           raw.TotalCost,
		PlanRows:            raw.PlanRows,
		PlanWidth:           raw.PlanWidth,
		ActualStartupTime:   raw.ActualStartupTime,
		ActualTotalTime:     totalTime,
		ActualRows:          raw.ActualRows * loops,
		ActualLoops:         loops,
		SharedHitBlocks:     raw.SharedHitBlocks,
		SharedReadBlocks:    raw.SharedReadBlocks,
		SharedDirtiedBlocks: raw.SharedDirtiedBlocks,
		SharedWrittenBlocks: raw.SharedWrittenBlocks,
		TempReadBlocks:      raw.TempReadBlocks,
		TempWrittenBlocks:   raw.TempWrittenBlocks,
		SortMethod:          raw.SortMethod,
		SortSpaceUsed:       raw.SortSpaceUsed,
		SortSpaceType:       raw.SortSpaceType,
		Filter:              raw.Filter,
		IndexCond:           raw.IndexCond,
		HashCond:            raw.HashCond,
		RecheckCond:         raw.RecheckCond,
		RowsRemovedByFilter: raw.RowsRemovedByFilter,
	}

	for i := range raw.Plans {
		child := convertNode(&raw.Plans[i], counter)
		node.Plans = append(node.Plans, child)
	}

	return node
}

func computeExclusiveTimes(node *model.ExplainPlanNode, totalPlanTime float64) {
	var childrenTime float64
	for _, child := range node.Plans {
		computeExclusiveTimes(child, totalPlanTime)
		childrenTime += child.ActualTotalTime
	}

	// Exclusive time is node's total time minus children
	node.ExclusiveTime = node.ActualTotalTime - childrenTime
	if node.ExclusiveTime < 0 {
		node.ExclusiveTime = 0
	}

	if totalPlanTime > 0 {
		node.PercentageOfTotal = math.Round((node.ExclusiveTime/totalPlanTime)*1000) / 10
	} else {
		node.PercentageOfTotal = 0
	}
}

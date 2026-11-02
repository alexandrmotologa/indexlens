package explain

import (
	"strings"
	"testing"
)

const sampleExplainJSON = `[
  {
    "Plan": {
      "Node Type": "Limit",
      "Startup Cost": 100.0,
      "Total Cost": 150.0,
      "Plan Rows": 10,
      "Plan Width": 32,
      "Actual Startup Time": 12.0,
      "Actual Total Time": 150.0,
      "Actual Rows": 10,
      "Actual Loops": 1,
      "Plans": [
        {
          "Node Type": "Sort",
          "Startup Cost": 100.0,
          "Total Cost": 140.0,
          "Plan Rows": 5000,
          "Plan Width": 32,
          "Actual Startup Time": 12.0,
          "Actual Total Time": 148.0,
          "Actual Rows": 10,
          "Actual Loops": 1,
          "Sort Method": "external merge Disk",
          "Sort Space Used": 16384,
          "Sort Space Type": "Disk",
          "Temp Written Blocks": 2048,
          "Plans": [
            {
              "Node Type": "Seq Scan",
              "Relation Name": "orders",
              "Schema": "public",
              "Startup Cost": 0.0,
              "Total Cost": 80.0,
              "Plan Rows": 5000,
              "Plan Width": 32,
              "Actual Startup Time": 0.5,
              "Actual Total Time": 120.0,
              "Actual Rows": 5200,
              "Actual Loops": 1,
              "Shared Hit Blocks": 1000,
              "Shared Read Blocks": 4000,
              "Filter": "(tenant_id = 42)",
              "Rows Removed by Filter": 95000
            }
          ]
        }
      ]
    },
    "Planning Time": 0.35,
    "Execution Time": 152.0
  }
]`

func TestParsePlanJSON(t *testing.T) {
	result, err := ParsePlanJSON([]byte(sampleExplainJSON))
	if err != nil {
		t.Fatalf("unexpected error parsing plan: %v", err)
	}

	if result.TotalExecutionTime != 152.0 {
		t.Errorf("expected execution time 152.0, got %f", result.TotalExecutionTime)
	}

	if result.RootNode == nil || result.RootNode.NodeType != "Limit" {
		t.Fatalf("expected root node to be Limit, got %+v", result.RootNode)
	}

	// Verify child hierarchy
	if len(result.RootNode.Plans) != 1 {
		t.Fatalf("expected 1 child for Limit, got %d", len(result.RootNode.Plans))
	}

	sortNode := result.RootNode.Plans[0]
	if sortNode.NodeType != "Sort" {
		t.Errorf("expected child to be Sort, got %s", sortNode.NodeType)
	}

	seqScanNode := sortNode.Plans[0]
	if seqScanNode.NodeType != "Seq Scan" || seqScanNode.RelationName != "orders" {
		t.Errorf("expected child to be Seq Scan on orders, got %s on %s", seqScanNode.NodeType, seqScanNode.RelationName)
	}

	// Verify bottleneck detection
	if len(result.Bottlenecks) == 0 {
		t.Fatalf("expected bottlenecks to be detected, got 0")
	}

	hasDiskSpill := false
	hasSeqScan := false
	for _, b := range result.Bottlenecks {
		if b.Type == "DISK_SPILL" {
			hasDiskSpill = true
		}
		if b.Type == "HEAVY_SEQ_SCAN" {
			hasSeqScan = true
		}
	}

	if !hasDiskSpill {
		t.Errorf("expected DISK_SPILL bottleneck to be detected")
	}
	if !hasSeqScan {
		t.Errorf("expected HEAVY_SEQ_SCAN bottleneck to be detected")
	}

	// Verify flamegraph bars
	bars := BuildFlamegraphStructure(result.RootNode, result.TotalExecutionTime)
	if len(bars) != 3 {
		t.Errorf("expected 3 flamegraph bars, got %d", len(bars))
	}
}

func TestParsePlanJSON_Invalid(t *testing.T) {
	_, err := ParsePlanJSON([]byte(""))
	if err == nil {
		t.Error("expected error for empty plan input")
	}

	_, err = ParsePlanJSON([]byte("not valid json"))
	if err == nil {
		t.Error("expected error for malformed json")
	}
}

func TestStaleStatsBottleneck(t *testing.T) {
	planJSON := `{
		"Plan": {
			"Node Type": "Seq Scan",
			"Relation Name": "large_table",
			"Plan Rows": 1,
			"Actual Rows": 50000,
			"Actual Total Time": 45.0,
			"Actual Loops": 1
		},
		"Execution Time": 46.0
	}`

	result, err := ParsePlanJSON([]byte(planJSON))
	if err != nil {
		t.Fatalf("failed to parse plan: %v", err)
	}

	foundStale := false
	for _, b := range result.Bottlenecks {
		if b.Type == "STALE_STATS" && strings.Contains(b.Message, "50000") {
			foundStale = true
			break
		}
	}

	if !foundStale {
		t.Errorf("expected STALE_STATS bottleneck for 50000x discrepancy")
	}
}

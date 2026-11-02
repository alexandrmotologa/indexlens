package model

import (
	"fmt"
	"time"
)

// Helper for formatting byte sizes.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetDemoQueries returns representative production slow queries.
func GetDemoQueries() []QueryStat {
	return []QueryStat{
		{
			QueryID:           "3489210984712",
			Query:             "SELECT * FROM orders WHERE tenant_id = $1 AND created_at >= $2 ORDER BY total_amount DESC LIMIT 50",
			Calls:             18450,
			TotalExecTime:     2435420.5,
			MeanExecTime:      132.0,
			MinExecTime:       12.4,
			MaxExecTime:       2840.1,
			StddevExecTime:    48.2,
			Rows:              922500,
			RowsPerCall:       50.0,
			SharedBlksHit:     1845000,
			SharedBlksRead:    984200,
			SharedBlksDirtied: 120,
			SharedBlksWritten: 45,
			TempBlksRead:      4200,
			TempBlksWritten:   4200,
			CacheHitRatio:     65.2,
		},
		{
			QueryID:           "8493021948201",
			Query:             "SELECT * FROM audit_logs WHERE user_id = $1 AND action = $2 AND created_at >= $3",
			Calls:             42100,
			TotalExecTime:     1894500.0,
			MeanExecTime:      45.0,
			MinExecTime:       5.1,
			MaxExecTime:       1240.5,
			StddevExecTime:    22.8,
			Rows:              1263000,
			RowsPerCall:       30.0,
			SharedBlksHit:     3450000,
			SharedBlksRead:    620000,
			SharedBlksDirtied: 0,
			SharedBlksWritten: 0,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     84.8,
		},
		{
			QueryID:           "1928301928401",
			Query:             "SELECT u.id, u.email, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id WHERE u.status = $1 GROUP BY u.id, u.email HAVING COUNT(o.id) > $2",
			Calls:             2400,
			TotalExecTime:     980400.0,
			MeanExecTime:      408.5,
			MinExecTime:       85.0,
			MaxExecTime:       4910.0,
			StddevExecTime:    98.4,
			Rows:              72000,
			RowsPerCall:       30.0,
			SharedBlksHit:     1250000,
			SharedBlksRead:    450000,
			SharedBlksDirtied: 20,
			SharedBlksWritten: 10,
			TempBlksRead:      18500,
			TempBlksWritten:   18500,
			CacheHitRatio:     73.5,
		},
		{
			QueryID:           "7192830192834",
			Query:             "UPDATE sessions SET last_seen_at = NOW(), ip_address = $1 WHERE session_token = $2",
			Calls:             350000,
			TotalExecTime:     875000.0,
			MeanExecTime:      2.5,
			MinExecTime:       0.3,
			MaxExecTime:       210.0,
			StddevExecTime:    4.1,
			Rows:              350000,
			RowsPerCall:       1.0,
			SharedBlksHit:     2800000,
			SharedBlksRead:    15000,
			SharedBlksDirtied: 350000,
			SharedBlksWritten: 82000,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     99.5,
		},
		{
			QueryID:           "5918273645102",
			Query:             "SELECT * FROM events WHERE payload->>'event_type' = $1 AND processed = false ORDER BY id ASC LIMIT 100",
			Calls:             85000,
			TotalExecTime:     765000.0,
			MeanExecTime:      9.0,
			MinExecTime:       1.2,
			MaxExecTime:       840.0,
			StddevExecTime:    11.2,
			Rows:              850000,
			RowsPerCall:       10.0,
			SharedBlksHit:     1900000,
			SharedBlksRead:    210000,
			SharedBlksDirtied: 50,
			SharedBlksWritten: 12,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     90.0,
		},
	}
}

// GetDemoTables returns representative tables with scan ratios and dead tuple counts.
func GetDemoTables() []TableStat {
	now := time.Now()
	t1 := now.Add(-14 * time.Minute)
	t2 := now.Add(-2 * time.Hour)
	t3 := now.Add(-6 * time.Hour)
	t4 := now.Add(-24 * time.Hour)

	return []TableStat{
		{
			SchemaName:      "public",
			TableName:       "orders",
			SeqScan:         45210,
			SeqTupRead:      38500000,
			IdxScan:         12400,
			IdxTupFetch:     48200,
			LiveTuples:      1250000,
			DeadTuples:      142000,
			DeadTupleRatio:  10.2,
			TableSizeBytes:  840 * 1024 * 1024,
			IndexSizeBytes:  180 * 1024 * 1024,
			TableSizePretty: FormatBytes(840 * 1024 * 1024),
			IndexSizePretty: FormatBytes(180 * 1024 * 1024),
			LastVacuum:      &t2,
			LastAutovacuum:  &t1,
		},
		{
			SchemaName:      "public",
			TableName:       "audit_logs",
			SeqScan:         78400,
			SeqTupRead:      94000000,
			IdxScan:         3200,
			IdxTupFetch:     15000,
			LiveTuples:      4800000,
			DeadTuples:      980000,
			DeadTupleRatio:  17.0,
			TableSizeBytes:  2400 * 1024 * 1024,
			IndexSizeBytes:  320 * 1024 * 1024,
			TableSizePretty: FormatBytes(2400 * 1024 * 1024),
			IndexSizePretty: FormatBytes(320 * 1024 * 1024),
			LastVacuum:      nil,
			LastAutovacuum:  &t3,
		},
		{
			SchemaName:      "public",
			TableName:       "sessions",
			SeqScan:         120,
			SeqTupRead:      45000,
			IdxScan:         890000,
			IdxTupFetch:     890000,
			LiveTuples:      350000,
			DeadTuples:      185000,
			DeadTupleRatio:  34.6,
			TableSizeBytes:  180 * 1024 * 1024,
			IndexSizeBytes:  65 * 1024 * 1024,
			TableSizePretty: FormatBytes(180 * 1024 * 1024),
			IndexSizePretty: FormatBytes(65 * 1024 * 1024),
			LastVacuum:      &t4,
			LastAutovacuum:  &t1,
		},
		{
			SchemaName:      "public",
			TableName:       "events",
			SeqScan:         62100,
			SeqTupRead:      24500000,
			IdxScan:         8400,
			IdxTupFetch:     28000,
			LiveTuples:      850000,
			DeadTuples:      240000,
			DeadTupleRatio:  22.0,
			TableSizeBytes:  650 * 1024 * 1024,
			IndexSizeBytes:  95 * 1024 * 1024,
			TableSizePretty: FormatBytes(650 * 1024 * 1024),
			IndexSizePretty: FormatBytes(95 * 1024 * 1024),
			LastVacuum:      nil,
			LastAutovacuum:  &t2,
		},
		{
			SchemaName:      "public",
			TableName:       "users",
			SeqScan:         450,
			SeqTupRead:      120000,
			IdxScan:         450000,
			IdxTupFetch:     450000,
			LiveTuples:      280000,
			DeadTuples:      4500,
			DeadTupleRatio:  1.6,
			TableSizeBytes:  120 * 1024 * 1024,
			IndexSizeBytes:  42 * 1024 * 1024,
			TableSizePretty: FormatBytes(120 * 1024 * 1024),
			IndexSizePretty: FormatBytes(42 * 1024 * 1024),
			LastVacuum:      &t3,
			LastAutovacuum:  &t3,
		},
	}
}

// GetDemoCacheStats returns cache hit metrics.
func GetDemoCacheStats() CacheHitStats {
	return CacheHitStats{
		GlobalHeapHitRatio:  96.4,
		GlobalIndexHitRatio: 98.9,
		OverallHitRatio:     97.2,
		TotalHeapHit:        18450000,
		TotalHeapRead:       690000,
		TotalIndexHit:       24800000,
		TotalIndexRead:      275000,
		TablesBelowThreshold: []TableCacheStat{
			{
				TableName:     "orders",
				HeapHitRatio:  65.2,
				IndexHitRatio: 94.1,
				HeapBlksHit:   1845000,
				HeapBlksRead:  984200,
				IndexBlksHit:  450000,
				IndexBlksRead: 28000,
			},
			{
				TableName:     "audit_logs",
				HeapHitRatio:  84.8,
				IndexHitRatio: 96.2,
				HeapBlksHit:   3450000,
				HeapBlksRead:  620000,
				IndexBlksHit:  820000,
				IndexBlksRead: 32000,
			},
		},
	}
}

// GetDemoExplainResult returns a realistic parsed EXPLAIN plan with flamegraph nodes.
func GetDemoExplainResult() ExplainAnalysisResult {
	root := &ExplainPlanNode{
		NodeID:            "node-1",
		NodeType:          "Limit",
		StartupCost:       1240.50,
		TotalCost:         1240.62,
		PlanRows:          50,
		PlanWidth:         148,
		ActualStartupTime: 182.4,
		ActualTotalTime:   182.6,
		ActualRows:        50,
		ActualLoops:       1,
		ExclusiveTime:     0.2,
		PercentageOfTotal: 0.1,
		Plans: []*ExplainPlanNode{
			{
				NodeID:            "node-2",
				NodeType:          "Sort",
				StartupCost:       1240.50,
				TotalCost:         1252.30,
				PlanRows:          4720,
				PlanWidth:         148,
				ActualStartupTime: 182.3,
				ActualTotalTime:   182.4,
				ActualRows:        50,
				ActualLoops:       1,
				ExclusiveTime:     34.2,
				PercentageOfTotal: 18.7,
				SortMethod:        "external merge Disk",
				SortSpaceUsed:     14336,
				SortSpaceType:     "Disk",
				TempReadBlocks:    1792,
				TempWrittenBlocks: 1792,
				Plans: []*ExplainPlanNode{
					{
						NodeID:              "node-3",
						NodeType:            "Seq Scan",
						RelationName:        "orders",
						Schema:              "public",
						Alias:               "orders",
						StartupCost:         0.00,
						TotalCost:           1120.40,
						PlanRows:            4720,
						PlanWidth:           148,
						ActualStartupTime:   0.8,
						ActualTotalTime:     148.2,
						ActualRows:          4890,
						ActualLoops:         1,
						ExclusiveTime:       148.2,
						PercentageOfTotal:   81.2,
						SharedHitBlocks:     12400,
						SharedReadBlocks:    8940,
						Filter:              "((tenant_id = 42) AND (created_at >= '2026-01-01 00:00:00'::timestamp))",
						RowsRemovedByFilter: 845110,
					},
				},
			},
		},
	}

	bottlenecks := []Bottleneck{
		{
			Type:           "HEAVY_SEQ_SCAN",
			Severity:       "CRITICAL",
			NodeID:         "node-3",
			NodeType:       "Seq Scan",
			Relation:       "orders",
			Message:        "Sequential scan read 850,000 tuples and discarded 845,110 rows (99.4% filter rejection)",
			Recommendation: "Create composite index on orders (tenant_id, created_at DESC) to convert Seq Scan to Index Scan",
			ImpactScore:    85.0,
		},
		{
			Type:           "DISK_SPILL",
			Severity:       "CRITICAL",
			NodeID:         "node-2",
			NodeType:       "Sort",
			Relation:       "orders",
			Message:        "Sort operation exceeded work_mem and spilled 14 MB to disk using external merge",
			Recommendation: "Increase work_mem or include sorted column in index: (tenant_id, created_at, total_amount DESC)",
			ImpactScore:    70.0,
		},
		{
			Type:           "CACHE_MISS",
			Severity:       "WARNING",
			NodeID:         "node-3",
			NodeType:       "Seq Scan",
			Relation:       "orders",
			Message:        "8,940 shared disk blocks read directly from storage (41.9% cache miss)",
			Recommendation: "Index will shrink accessed working set into shared_buffers",
			ImpactScore:    55.0,
		},
	}

	var flat []*ExplainPlanNode
	var collect func(n *ExplainPlanNode)
	collect = func(n *ExplainPlanNode) {
		flat = append(flat, n)
		for _, child := range n.Plans {
			collect(child)
		}
	}
	collect(root)

	return ExplainAnalysisResult{
		RootNode:           root,
		TotalExecutionTime: 182.6,
		PlanningTime:       1.45,
		TotalBuffersHit:    12400,
		TotalBuffersRead:   8940,
		CacheHitRatio:      58.1,
		Bottlenecks:        bottlenecks,
		FlattenedNodes:     flat,
	}
}

// GetDemoRecommendations returns synthesized missing index suggestions.
func GetDemoRecommendations() []IndexRecommendation {
	return []IndexRecommendation{
		{
			ID:                     "rec-orders-tenant-created",
			TableName:              "orders",
			IndexName:              "idx_orders_tenant_created",
			Columns:                []string{"tenant_id", "created_at DESC"},
			Predicate:              "",
			Reason:                 "45,210 sequential scans reading 38.5M tuples. Query filters by equality on tenant_id and range on created_at.",
			DDL:                    "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_tenant_created ON public.orders (tenant_id, created_at DESC);",
			EstimatedScanReduction: 88.5,
			SeqScanCount:           45210,
			SeqTuplesRead:          38500000,
			Priority:               "HIGH",
		},
		{
			ID:                     "rec-audit-user-action",
			TableName:              "audit_logs",
			IndexName:              "idx_audit_logs_user_action_created",
			Columns:                []string{"user_id", "action", "created_at DESC"},
			Predicate:              "",
			Reason:                 "78,400 sequential scans reading 94M tuples. High frequency query filters on user_id and action.",
			DDL:                    "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_action_created ON public.audit_logs (user_id, action, created_at DESC);",
			EstimatedScanReduction: 94.2,
			SeqScanCount:           78400,
			SeqTuplesRead:          94000000,
			Priority:               "HIGH",
		},
		{
			ID:                     "rec-events-unprocessed",
			TableName:              "events",
			IndexName:              "idx_events_unprocessed_type",
			Columns:                []string{"((payload->>'event_type'))", "id ASC"},
			Predicate:              "processed = false",
			Reason:                 "Partial index opportunity: 62,100 sequential scans fetching pending event queue rows.",
			DDL:                    "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_events_unprocessed_type ON public.events (((payload->>'event_type')), id ASC) WHERE processed = false;",
			EstimatedScanReduction: 96.0,
			SeqScanCount:           62100,
			SeqTuplesRead:          24500000,
			Priority:               "MEDIUM",
		},
	}
}

// GetDemoBloat returns estimated table bloat metrics.
func GetDemoBloat() []BloatInfo {
	return []BloatInfo{
		{
			TableName:      "audit_logs",
			SizeBytes:      2400 * 1024 * 1024,
			BloatBytes:     1050 * 1024 * 1024,
			BloatRatio:     43.8,
			LiveTuples:     4800000,
			DeadTuples:     980000,
			DeadTupleRatio: 17.0,
			SizePretty:     FormatBytes(2400 * 1024 * 1024),
			BloatPretty:    FormatBytes(1050 * 1024 * 1024),
		},
		{
			TableName:      "sessions",
			SizeBytes:      180 * 1024 * 1024,
			BloatBytes:     78 * 1024 * 1024,
			BloatRatio:     43.3,
			LiveTuples:     350000,
			DeadTuples:     185000,
			DeadTupleRatio: 34.6,
			SizePretty:     FormatBytes(180 * 1024 * 1024),
			BloatPretty:    FormatBytes(78 * 1024 * 1024),
		},
		{
			TableName:      "events",
			SizeBytes:      650 * 1024 * 1024,
			BloatBytes:     195 * 1024 * 1024,
			BloatRatio:     30.0,
			LiveTuples:     850000,
			DeadTuples:     240000,
			DeadTupleRatio: 22.0,
			SizePretty:     FormatBytes(650 * 1024 * 1024),
			BloatPretty:    FormatBytes(195 * 1024 * 1024),
		},
		{
			TableName:      "orders",
			SizeBytes:      840 * 1024 * 1024,
			BloatBytes:     125 * 1024 * 1024,
			BloatRatio:     14.9,
			LiveTuples:     1250000,
			DeadTuples:     142000,
			DeadTupleRatio: 10.2,
			SizePretty:     FormatBytes(840 * 1024 * 1024),
			BloatPretty:    FormatBytes(125 * 1024 * 1024),
		},
	}
}

// GetDemoAutovacuum returns active autovacuum workers and wraparound metrics.
func GetDemoAutovacuum() ([]AutovacuumWorker, WraparoundInfo) {
	workers := []AutovacuumWorker{
		{
			PID:                24912,
			Database:           "production_db",
			Table:              "audit_logs",
			Phase:              "vacuuming indexes",
			HeapBlksTotal:      307200,
			HeapBlksScanned:    307200,
			HeapBlksVacuumed:   215000,
			IndexVacuumCount:   1,
			ProgressPercentage: 70.0,
		},
		{
			PID:                24988,
			Database:           "production_db",
			Table:              "sessions",
			Phase:              "scanning heap",
			HeapBlksTotal:      23040,
			HeapBlksScanned:    9800,
			HeapBlksVacuumed:   0,
			IndexVacuumCount:   0,
			ProgressPercentage: 42.5,
		},
	}

	wraparound := WraparoundInfo{
		DatabaseName:    "production_db",
		CurrentTxidAge:  54200100,
		MaxTxidAge:      200000000,
		PercentConsumed: 27.1,
		RemainingTxids:  145799900,
		Status:          "HEALTHY",
	}

	return workers, wraparound
}

// GetDemoOverview returns system overview metrics for the demo state.
func GetDemoOverview() SystemOverview {
	return SystemOverview{
		Connected:               true,
		DemoMode:                true,
		DatabaseVersion:         "PostgreSQL 16.4 on x86_64-pc-linux-gnu",
		DatabaseName:            "production_db (simulated)",
		ActiveConnections:       48,
		MaxConnections:          200,
		BufferCacheHitRatio:     97.2,
		TotalQueriesTracked:     5,
		UnderIndexedTablesCount: 3,
		BloatWarningCount:       2,
		WraparoundSafetyPercent: 72.9,
		ActiveAutovacuums:       2,
	}
}

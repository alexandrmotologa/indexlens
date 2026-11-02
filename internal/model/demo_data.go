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
			P50ExecTime:       122.4,
			P90ExecTime:       193.7,
			P95ExecTime:       211.3,
			P99ExecTime:       244.2,
			JitterScore:       0.365,
			JitterLevel:       "MODERATE",
			Rows:              922500,
			RowsPerCall:       50.0,
			SharedBlksHit:     1845000,
			SharedBlksRead:    984200,
			SharedBlksDirtied: 120,
			SharedBlksWritten: 45,
			TempBlksRead:      4200,
			TempBlksWritten:   4200,
			CacheHitRatio:     65.2,
			AntiPatterns:      nil,
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
			P50ExecTime:       40.4,
			P90ExecTime:       74.2,
			P95ExecTime:       82.5,
			P99ExecTime:       98.0,
			JitterScore:       0.507,
			JitterLevel:       "MODERATE",
			Rows:              1263000,
			RowsPerCall:       30.0,
			SharedBlksHit:     3450000,
			SharedBlksRead:    620000,
			SharedBlksDirtied: 0,
			SharedBlksWritten: 0,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     84.8,
			AntiPatterns:      nil,
		},
		{
			QueryID:           "1928301928401",
			Query:             "SELECT u.id, u.email, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id WHERE u.status = $1 GROUP BY u.id, u.email HAVING COUNT(o.id) > $2",
			Calls:             2400,
			TotalExecTime:     980400.0,
			MeanExecTime:      408.5,
			MinExecTime:       85.0,
			MaxExecTime:       4910.0,
			StddevExecTime:    340.2,
			P50ExecTime:       340.5,
			P90ExecTime:       844.0,
			P95ExecTime:       968.1,
			P99ExecTime:       1199.7,
			JitterScore:       0.833,
			JitterLevel:       "SPIKY",
			Rows:              72000,
			RowsPerCall:       30.0,
			SharedBlksHit:     1250000,
			SharedBlksRead:    450000,
			SharedBlksDirtied: 20,
			SharedBlksWritten: 10,
			TempBlksRead:      18500,
			TempBlksWritten:   18500,
			CacheHitRatio:     73.5,
			AntiPatterns:      nil,
		},
		{
			QueryID:           "7192830192834",
			Query:             "UPDATE sessions SET last_seen_at = NOW(), ip_address = $1 WHERE session_token = $2",
			Calls:             350000,
			TotalExecTime:     875000.0,
			MeanExecTime:      2.5,
			MinExecTime:       0.3,
			MaxExecTime:       210.0,
			StddevExecTime:    0.4,
			P50ExecTime:       2.4,
			P90ExecTime:       3.0,
			P95ExecTime:       3.2,
			P99ExecTime:       3.4,
			JitterScore:       0.160,
			JitterLevel:       "STABLE",
			Rows:              350000,
			RowsPerCall:       1.0,
			SharedBlksHit:     2800000,
			SharedBlksRead:    15000,
			SharedBlksDirtied: 350000,
			SharedBlksWritten: 82000,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     99.5,
			AntiPatterns:      nil,
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
			P50ExecTime:       6.8,
			P90ExecTime:       23.3,
			P95ExecTime:       27.4,
			P99ExecTime:       35.1,
			JitterScore:       1.244,
			JitterLevel:       "SPIKY",
			Rows:              850000,
			RowsPerCall:       10.0,
			SharedBlksHit:     1900000,
			SharedBlksRead:    210000,
			SharedBlksDirtied: 50,
			SharedBlksWritten: 12,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     90.0,
			AntiPatterns:      nil,
		},
		{
			QueryID:           "9102837461928",
			Query:             "SELECT id, email, created_at FROM users WHERE lower(email) = $1 LIMIT 1",
			Calls:             64000,
			TotalExecTime:     640000.0,
			MeanExecTime:      10.0,
			MinExecTime:       0.8,
			MaxExecTime:       420.0,
			StddevExecTime:    9.4,
			P50ExecTime:       8.1,
			P90ExecTime:       22.0,
			P95ExecTime:       25.5,
			P99ExecTime:       31.9,
			JitterScore:       0.940,
			JitterLevel:       "SPIKY",
			Rows:              64000,
			RowsPerCall:       1.0,
			SharedBlksHit:     320000,
			SharedBlksRead:    180000,
			SharedBlksDirtied: 0,
			SharedBlksWritten: 0,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     64.0,
			AntiPatterns: []SQLAntiPattern{
				{
					Type:          "FUNCTION_ON_INDEX_COLUMN",
					Severity:      "CRITICAL",
					Snippet:       "lower(email) =",
					Message:       "Calling function 'lower' on column 'email' invalidates standard B-Tree index lookups and causes full sequential scans.",
					RewriteAdvice: "Either store normalized values in 'email' or create an expression index to accelerate this filter.",
					SuggestedDDL:  "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expr_users_email ON users (lower(email));",
				},
			},
		},
		{
			QueryID:           "4819203948172",
			Query:             "SELECT id, name, sku FROM products WHERE name ILIKE '%phone%' ORDER BY id ASC LIMIT 50 OFFSET 15000",
			Calls:             12500,
			TotalExecTime:     562500.0,
			MeanExecTime:      45.0,
			MinExecTime:       4.5,
			MaxExecTime:       920.0,
			StddevExecTime:    38.5,
			P50ExecTime:       37.3,
			P90ExecTime:       94.3,
			P95ExecTime:       108.3,
			P99ExecTime:       134.5,
			JitterScore:       0.855,
			JitterLevel:       "SPIKY",
			Rows:              625000,
			RowsPerCall:       50.0,
			SharedBlksHit:     410000,
			SharedBlksRead:    220000,
			SharedBlksDirtied: 0,
			SharedBlksWritten: 0,
			TempBlksRead:      0,
			TempBlksWritten:   0,
			CacheHitRatio:     65.1,
			AntiPatterns: []SQLAntiPattern{
				{
					Type:          "LEADING_WILDCARD",
					Severity:      "CRITICAL",
					Snippet:       "name ILIKE '%phone%'",
					Message:       "Leading wildcard pattern '%phone%' prevents B-Tree bisection, forcing table scans or expensive full index scans.",
					RewriteAdvice: "For substring and suffix matching, utilize PostgreSQL's pg_trgm extension with a GIN index instead of plain LIKE.",
					SuggestedDDL:  "CREATE EXTENSION IF NOT EXISTS pg_trgm;\nCREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gin_products_name ON products USING gin (name gin_trgm_ops);",
				},
				{
					Type:          "DEEP_OFFSET",
					Severity:      "WARNING",
					Snippet:       "OFFSET 15000",
					Message:       "Deep OFFSET of 15000 rows forces PostgreSQL to read and discard all prior rows, causing severe buffer churn and disk IO.",
					RewriteAdvice: "Switch from numeric OFFSET to Keyset (Seek) Pagination using indexed cursor columns (e.g. WHERE id > :last_seen_id ORDER BY id ASC LIMIT :page_size).",
				},
			},
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

// GetDemoRecommendations returns synthesized missing index suggestions with estimated size.
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
			EstimatedSizeBytes:     48 * 1024 * 1024,
			EstimatedSizePretty:    "48.0 MB",
			WritePenalty:           "LOW",
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
			EstimatedSizeBytes:     142 * 1024 * 1024,
			EstimatedSizePretty:    "142.0 MB",
			WritePenalty:           "MODERATE",
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
			EstimatedSizeBytes:     12 * 1024 * 1024,
			EstimatedSizePretty:    "12.0 MB",
			WritePenalty:           "LOW",
			SeqScanCount:           62100,
			SeqTuplesRead:          24500000,
			Priority:               "MEDIUM",
		},
	}
}

// GetDemoUnusedIndexes returns zero-scan and redundant indexes with drop DDL.
func GetDemoUnusedIndexes() []UnusedIndex {
	return []UnusedIndex{
		{
			SchemaName:      "public",
			TableName:       "orders",
			IndexName:       "idx_orders_legacy_status",
			IndexSizeBytes:  420 * 1024 * 1024,
			IndexSizePretty: "420.0 MB",
			IndexScans:      0,
			TuplesRead:      0,
			TuplesFetched:   0,
			IsRedundant:     false,
			Columns:         []string{"legacy_status"},
			DropDDL:         "DROP INDEX CONCURRENTLY IF EXISTS public.idx_orders_legacy_status;",
			Reason:          "0 scans recorded across 18,450 queries. Index consumes 420 MB of disk and shared_buffers memory.",
			Priority:        "HIGH",
		},
		{
			SchemaName:      "public",
			TableName:       "audit_logs",
			IndexName:       "idx_audit_logs_user_id_only",
			IndexSizeBytes:  310 * 1024 * 1024,
			IndexSizePretty: "310.0 MB",
			IndexScans:      14,
			TuplesRead:      42,
			TuplesFetched:   14,
			IsRedundant:     true,
			RedundantWith:   "idx_audit_logs_user_action_created (user_id, action, created_at)",
			Columns:         []string{"user_id"},
			DropDDL:         "DROP INDEX CONCURRENTLY IF EXISTS public.idx_audit_logs_user_id_only;",
			Reason:          "Redundant index. The composite index on (user_id, action, created_at) already covers leading prefix queries.",
			Priority:        "HIGH",
		},
		{
			SchemaName:      "public",
			TableName:       "events",
			IndexName:       "idx_events_created_at",
			IndexSizeBytes:  95 * 1024 * 1024,
			IndexSizePretty: "95.0 MB",
			IndexScans:      2,
			TuplesRead:      12,
			TuplesFetched:   2,
			IsRedundant:     false,
			Columns:         []string{"created_at"},
			DropDDL:         "DROP INDEX CONCURRENTLY IF EXISTS public.idx_events_created_at;",
			Reason:          "Negligible scan count (2 scans). Table updates are bearing index maintenance write overhead for unused index.",
			Priority:        "MEDIUM",
		},
	}
}

// GetDemoLockContention returns a realistic lock graph and blocking process tree.
func GetDemoLockContention() LockContentionTree {
	rootBlocker := LiveActivityItem{
		PID:           18402,
		User:          "app_worker",
		Database:      "production_db",
		State:         "idle in transaction",
		DurationMs:    14250.0,
		WaitEventType: "Client",
		WaitEvent:     "ClientRead",
		Query:         "UPDATE orders SET status = 'processing' WHERE id = 984021;",
		BlockedPIDs:   []int64{18455, 18460},
	}

	blocked1 := LiveActivityItem{
		PID:           18455,
		User:          "web_api",
		Database:      "production_db",
		State:         "active",
		DurationMs:    12840.0,
		WaitEventType: "Lock",
		WaitEvent:     "transactionid",
		Query:         "SELECT * FROM orders WHERE id = 984021 FOR UPDATE;",
		BlockingPID:   18402,
	}

	blocked2 := LiveActivityItem{
		PID:           18460,
		User:          "billing_svc",
		Database:      "production_db",
		State:         "active",
		DurationMs:    8910.0,
		WaitEventType: "Lock",
		WaitEvent:     "tuple",
		Query:         "UPDATE orders SET total_amount = 499.00 WHERE id = 984021;",
		BlockingPID:   18402,
	}

	return LockContentionTree{
		RootBlockers:        []LiveActivityItem{rootBlocker},
		ActiveProcesses:     []LiveActivityItem{rootBlocker, blocked1, blocked2},
		TotalBlockedQueries: 2,
		ActiveLocksCount:    14,
		HasContention:       true,
	}
}

// GetDemoPlanDiff returns a representative comparison between unindexed vs indexed query runs.
func GetDemoPlanDiff() PlanDiffResult {
	return PlanDiffResult{
		BaselineExecutionTime: 182.6,
		CurrentExecutionTime:  4.2,
		TimeDeltaMs:           -178.4,
		TimeDeltaPct:          -97.7,
		BaselineBuffersHit:    12400,
		CurrentBuffersHit:     820,
		BaselineBuffersRead:   8940,
		CurrentBuffersRead:    0,
		Verdict:               "IMPROVED",
		Summary:               "Execution time dropped by 97.7% (182.6ms -> 4.2ms). Sequential scan converted to B-Tree Index Scan, eliminating 8,940 disk block reads and external sort spill.",
		NodeChanges: []PlanNodeDiff{
			{
				Path:          "root.sort.scan",
				BaselineType:  "Seq Scan on orders (148.2 ms)",
				CurrentType:   "Index Scan using idx_orders_tenant_created (3.1 ms)",
				BaselineTime:  148.2,
				CurrentTime:   3.1,
				TimeDeltaMs:   -145.1,
				ChangeSummary: "Table scan replaced by index search; discarded row overhead eliminated.",
			},
			{
				Path:          "root.sort",
				BaselineType:  "Sort (external merge Disk 14 MB)",
				CurrentType:   "Eliminated (Index-Organized Sorting)",
				BaselineTime:  34.2,
				CurrentTime:   0.0,
				TimeDeltaMs:   -34.2,
				ChangeSummary: "Disk sort spill completely bypassed by ordered B-Tree traversal.",
			},
		},
	}
}

// GetDemoHypoSimulation returns simulated cost impact.
func GetDemoHypoSimulation() HypoSimulationResult {
	return HypoSimulationResult{
		Supported:          true,
		ExtensionInstalled: true,
		OriginalCost:       1240.62,
		SimulatedCost:      14.25,
		CostReductionPct:   98.85,
		OriginalPlanType:   "Seq Scan + Sort (Spill)",
		SimulatedPlanType:  "Index Scan using <hypo_idx_orders_tenant_created>",
		IndexName:          "idx_orders_tenant_created",
		Columns:            "tenant_id, created_at DESC",
		Message:            "Planner verified! PostgreSQL query planner selects the hypothetical index, cutting plan cost from 1240.6 to 14.3.",
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
		TotalQueriesTracked:     7,
		UnderIndexedTablesCount: 3,
		UnusedIndexesCount:      3,
		ActiveLocksCount:        14,
		BloatWarningCount:       2,
		WraparoundSafetyPercent: 72.9,
		ActiveAutovacuums:       2,
	}
}

// GetDemoTuningReport returns realistic configuration tuning recommendations.
func GetDemoTuningReport(profile TuningProfile) TuningReport {
	if profile.TotalRAMGB <= 0 {
		profile.TotalRAMGB = 32
	}
	if profile.VCPUs <= 0 {
		profile.VCPUs = 8
	}
	if profile.DiskType == "" {
		profile.DiskType = "NVMe"
	}
	if profile.WorkloadType == "" {
		profile.WorkloadType = "OLTP"
	}

	settings := []ConfigSetting{
		{
			Name:             "shared_buffers",
			CurrentValue:     "128MB",
			RecommendedValue: "8192MB",
			Unit:             "MB",
			Category:         "Memory",
			RestartRequired:  true,
			Reason:           "Allocates 25% of system RAM to PostgreSQL dedicated page cache, preventing kernel context-switch stalls.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET shared_buffers = '8192MB';",
		},
		{
			Name:             "effective_cache_size",
			CurrentValue:     "4096MB",
			RecommendedValue: "24576MB",
			Unit:             "MB",
			Category:         "Memory",
			RestartRequired:  false,
			Reason:           "Planner assumption for total memory available for disk caching (PostgreSQL buffer cache + OS page cache).",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET effective_cache_size = '24576MB';",
		},
		{
			Name:             "work_mem",
			CurrentValue:     "4MB",
			RecommendedValue: "81MB",
			Unit:             "MB",
			Category:         "Memory",
			RestartRequired:  false,
			Reason:           "Allocated per sort/hash operation before spilling temporary files to disk. Sized for concurrency without OOM risk.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET work_mem = '81MB';",
		},
		{
			Name:             "maintenance_work_mem",
			CurrentValue:     "64MB",
			RecommendedValue: "2048MB",
			Unit:             "MB",
			Category:         "Memory",
			RestartRequired:  false,
			Reason:           "Accelerates VACUUM, CREATE INDEX, and ALTER TABLE foreign key validations.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET maintenance_work_mem = '2048MB';",
		},
		{
			Name:             "checkpoint_completion_target",
			CurrentValue:     "0.5",
			RecommendedValue: "0.9",
			Unit:             "ratio",
			Category:         "WAL",
			RestartRequired:  false,
			Reason:           "Spreads dirty page write I/O across 90% of the checkpoint duration, eliminating I/O spikes.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET checkpoint_completion_target = '0.9';",
		},
		{
			Name:             "wal_buffers",
			CurrentValue:     "4MB",
			RecommendedValue: "64MB",
			Unit:             "MB",
			Category:         "WAL",
			RestartRequired:  true,
			Reason:           "Dedicates buffer space for unwritten WAL data, avoiding disk write bottlenecks during concurrent transactions.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET wal_buffers = '64MB';",
		},
		{
			Name:             "random_page_cost",
			CurrentValue:     "4.0",
			RecommendedValue: "1.1",
			Unit:             "ratio",
			Category:         "Query Planner",
			RestartRequired:  false,
			Reason:           "Cost estimate for non-sequential disk reads. 1.1 reflects fast flash/NVMe storage random access.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET random_page_cost = '1.1';",
		},
		{
			Name:             "effective_io_concurrency",
			CurrentValue:     "1",
			RecommendedValue: "200",
			Unit:             "requests",
			Category:         "Query Planner",
			RestartRequired:  false,
			Reason:           "Allows asynchronous prefetching of heap pages during bitmap index scans.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET effective_io_concurrency = '200';",
		},
		{
			Name:             "max_worker_processes",
			CurrentValue:     "8",
			RecommendedValue: "8",
			Unit:             "workers",
			Category:         "Parallelism",
			RestartRequired:  true,
			Reason:           "Caps background workers to match system physical cores for optimal parallel query execution.",
			Status:           "OPTIMAL",
			AlterDDL:         "ALTER SYSTEM SET max_worker_processes = '8';",
		},
		{
			Name:             "max_parallel_workers_per_gather",
			CurrentValue:     "2",
			RecommendedValue: "4",
			Unit:             "workers",
			Category:         "Parallelism",
			RestartRequired:  false,
			Reason:           "Maximum parallel threads spawned for a single parallel query node.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET max_parallel_workers_per_gather = '4';",
		},
		{
			Name:             "max_parallel_maintenance_workers",
			CurrentValue:     "2",
			RecommendedValue: "4",
			Unit:             "workers",
			Category:         "Parallelism",
			RestartRequired:  false,
			Reason:           "Parallel threads for CREATE INDEX and autovacuum operations.",
			Status:           "ATTENTION",
			AlterDDL:         "ALTER SYSTEM SET max_parallel_maintenance_workers = '4';",
		},
	}

	alterScript := `-- IndexLens Generated PostgreSQL Hardware Tuning Script
-- Profile: 32 GB RAM, 8 vCPUs, NVMe storage, OLTP workload

ALTER SYSTEM SET shared_buffers = '8192MB';
ALTER SYSTEM SET effective_cache_size = '24576MB';
ALTER SYSTEM SET work_mem = '81MB';
ALTER SYSTEM SET maintenance_work_mem = '2048MB';
ALTER SYSTEM SET checkpoint_completion_target = '0.9';
ALTER SYSTEM SET wal_buffers = '64MB';
ALTER SYSTEM SET random_page_cost = '1.1';
ALTER SYSTEM SET effective_io_concurrency = '200';
ALTER SYSTEM SET max_worker_processes = '8';
ALTER SYSTEM SET max_parallel_workers_per_gather = '4';
ALTER SYSTEM SET max_parallel_maintenance_workers = '4';

-- Apply reloadable parameters:
SELECT pg_reload_conf();
`

	return TuningReport{
		Profile:        profile,
		Settings:       settings,
		OptimalCount:   1,
		AttentionCount: 10,
		AlterScript:    alterScript,
	}
}

// GetDemoSchemaHealth returns sample unindexed foreign keys, missing primary keys, and TOAST bloat.
func GetDemoSchemaHealth() SchemaHealthReport {
	return SchemaHealthReport{
		UnindexedFKs: []UnindexedForeignKey{
			{
				ConstraintName: "fk_order_items_order_id",
				ChildTable:     "order_items",
				ParentTable:    "orders",
				Columns:        []string{"order_id"},
				SuggestedDDL:   "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fk_order_items_order_id ON order_items (order_id);",
				Severity:       "CRITICAL",
				Reason:         "Lacks supporting index on child table 'order_items'. Modifications or deletes on parent table 'orders' will acquire ShareRowExclusiveLock and trigger full sequential scans.",
			},
			{
				ConstraintName: "fk_audit_logs_user_id",
				ChildTable:     "audit_logs",
				ParentTable:    "users",
				Columns:        []string{"user_id"},
				SuggestedDDL:   "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fk_audit_logs_user_id ON audit_logs (user_id);",
				Severity:       "CRITICAL",
				Reason:         "Lacks supporting index on child table 'audit_logs'. Deleting user records will perform a sequential scan across 4.8M audit log tuples.",
			},
			{
				ConstraintName: "fk_payments_subscription_id",
				ChildTable:     "payments",
				ParentTable:    "subscriptions",
				Columns:        []string{"subscription_id"},
				SuggestedDDL:   "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fk_payments_subscription_id ON payments (subscription_id);",
				Severity:       "CRITICAL",
				Reason:         "Foreign key without index. Concurrent subscription billing updates experience lock queues.",
			},
		},
		MissingPKs: []TableMissingPK{
			{
				SchemaName:      "public",
				TableName:       "event_logs",
				EstimatedRows:   1450000,
				TotalSizePretty: "420 MB",
				SuggestedDDL:    "ALTER TABLE public.event_logs ADD PRIMARY KEY (id);",
			},
			{
				SchemaName:      "public",
				TableName:       "cache_store",
				EstimatedRows:   85000,
				TotalSizePretty: "28 MB",
				SuggestedDDL:    "ALTER TABLE public.cache_store ADD PRIMARY KEY (key);",
			},
		},
		ToastRelations: []ToastRelationInfo{
			{
				TableName:       "user_profiles",
				ToastTableName:  "pg_toast_16421",
				MainSizeBytes:   120 * 1024 * 1024,
				ToastSizeBytes:  840 * 1024 * 1024,
				MainSizePretty:  "120 MB",
				ToastSizePretty: "840 MB",
				ToastRatio:      87.5,
				Recommendation:  "High out-of-line storage ratio (87.5%). Wide JSONB documents cause detached disk seeks. Consider enabling LZ4 column compression (ALTER TABLE user_profiles SET COMPRESSION lz4).",
			},
			{
				TableName:       "raw_documents",
				ToastTableName:  "pg_toast_16590",
				MainSizeBytes:   310 * 1024 * 1024,
				ToastSizeBytes:  2100 * 1024 * 1024,
				MainSizePretty:  "310 MB",
				ToastSizePretty: "2.1 GB",
				ToastRatio:      87.1,
				Recommendation:  "High out-of-line storage ratio. Consider offloading binary files or raw markdown payloads to object storage.",
			},
		},
		TotalIssues: 5,
		HealthScore: 55,
	}
}

// GetDemoConnectionPoolHealth returns sample connection saturation and PgBouncer recommendations.
func GetDemoConnectionPoolHealth() ConnectionPoolHealth {
	return ConnectionPoolHealth{
		TotalConnections:    142,
		MaxConnections:      200,
		UtilizationPercent:  71.0,
		ActiveCount:         38,
		IdleCount:           94,
		IdleInTxCount:       10,
		MaxIdleInTxDuration: 840.0,
		StateBreakdown: map[string]int{
			"active":              38,
			"idle":                94,
			"idle in transaction": 10,
		},
		OptimalPoolSize: 17,
		Status:          "WARNING",
		Warnings: []string{
			"Moderate connection saturation (71.0% of max_connections).",
			"Detected 10 connection(s) 'idle in transaction' for up to 840 seconds. Long-lived transactions block vacuum tuple cleanup and risk table bloat.",
		},
		PgBouncerAdvice: "Current client connection count (142) significantly exceeds hardware sweet spot (17). Deploy PgBouncer in transaction pooling mode to cap server backends at 17 and slash CPU context switching.",
	}
}

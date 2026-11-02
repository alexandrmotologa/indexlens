package model

import "time"

// QueryStat represents execution statistics for a normalized query from pg_stat_statements.
type QueryStat struct {
	QueryID           string  `json:"queryId"`
	Query             string  `json:"query"`
	Calls             int64   `json:"calls"`
	TotalExecTime     float64 `json:"totalExecTime"` // milliseconds
	MeanExecTime      float64 `json:"meanExecTime"`  // milliseconds
	MinExecTime       float64 `json:"minExecTime"`   // milliseconds
	MaxExecTime       float64 `json:"maxExecTime"`   // milliseconds
	StddevExecTime    float64 `json:"stddevExecTime"`
	Rows              int64   `json:"rows"`
	RowsPerCall       float64 `json:"rowsPerCall"`
	SharedBlksHit     int64   `json:"sharedBlksHit"`
	SharedBlksRead    int64   `json:"sharedBlksRead"`
	SharedBlksDirtied int64   `json:"sharedBlksDirtied"`
	SharedBlksWritten int64   `json:"sharedBlksWritten"`
	LocalBlksHit      int64   `json:"localBlksHit"`
	LocalBlksRead     int64   `json:"localBlksRead"`
	TempBlksRead      int64   `json:"tempBlksRead"`
	TempBlksWritten   int64   `json:"tempBlksWritten"`
	CacheHitRatio     float64          `json:"cacheHitRatio"` // percentage (0.0 to 100.0)
	P50ExecTime       float64          `json:"p50ExecTime"`   // ms
	P90ExecTime       float64          `json:"p90ExecTime"`   // ms
	P95ExecTime       float64          `json:"p95ExecTime"`   // ms
	P99ExecTime       float64          `json:"p99ExecTime"`   // ms
	JitterScore       float64          `json:"jitterScore"`   // stddev / mean
	JitterLevel       string           `json:"jitterLevel"`   // STABLE, MODERATE, SPIKY
	AntiPatterns      []SQLAntiPattern `json:"antiPatterns"`
}

// TableStat captures sequential vs index scans, tuple counts, and bloat signals.
type TableStat struct {
	SchemaName      string     `json:"schemaName"`
	TableName       string     `json:"tableName"`
	SeqScan         int64      `json:"seqScan"`
	SeqTupRead      int64      `json:"seqTupRead"`
	IdxScan         int64      `json:"idxScan"`
	IdxTupFetch     int64      `json:"idxTupFetch"`
	LiveTuples      int64      `json:"liveTuples"`
	DeadTuples      int64      `json:"deadTuples"`
	DeadTupleRatio  float64    `json:"deadTupleRatio"` // percentage
	TableSizeBytes  int64      `json:"tableSizeBytes"`
	IndexSizeBytes  int64      `json:"indexSizeBytes"`
	TableSizePretty string     `json:"tableSizePretty"`
	IndexSizePretty string     `json:"indexSizePretty"`
	LastVacuum      *time.Time `json:"lastVacuum"`
	LastAutovacuum  *time.Time `json:"lastAutovacuum"`
}

// TableCacheStat represents cache efficiency for a single relation.
type TableCacheStat struct {
	TableName     string  `json:"tableName"`
	HeapBlksHit   int64   `json:"heapBlksHit"`
	HeapBlksRead  int64   `json:"heapBlksRead"`
	HeapHitRatio  float64 `json:"heapHitRatio"`
	IndexBlksHit  int64   `json:"indexBlksHit"`
	IndexBlksRead int64   `json:"indexBlksRead"`
	IndexHitRatio float64 `json:"indexHitRatio"`
}

// CacheHitStats captures global and per-table buffer hit ratios.
type CacheHitStats struct {
	GlobalHeapHitRatio   float64          `json:"globalHeapHitRatio"`
	GlobalIndexHitRatio  float64          `json:"globalIndexHitRatio"`
	OverallHitRatio      float64          `json:"overallHitRatio"`
	TotalHeapHit         int64            `json:"totalHeapHit"`
	TotalHeapRead        int64            `json:"totalHeapRead"`
	TotalIndexHit        int64            `json:"totalIndexHit"`
	TotalIndexRead       int64            `json:"totalIndexRead"`
	TablesBelowThreshold []TableCacheStat `json:"tablesBelowThreshold"`
}

// ExplainPlanNode is the recursive AST node for EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON).
type ExplainPlanNode struct {
	NodeID              string             `json:"nodeId"`
	NodeType            string             `json:"nodeType"`
	RelationName        string             `json:"relationName,omitempty"`
	Schema              string             `json:"schema,omitempty"`
	Alias               string             `json:"alias,omitempty"`
	IndexName           string             `json:"indexName,omitempty"`
	StartupCost         float64            `json:"startupCost"`
	TotalCost           float64            `json:"totalCost"`
	PlanRows            float64            `json:"planRows"`
	PlanWidth           int                `json:"planWidth"`
	ActualStartupTime   float64            `json:"actualStartupTime"` // ms
	ActualTotalTime     float64            `json:"actualTotalTime"`   // ms
	ActualRows          float64            `json:"actualRows"`
	ActualLoops         float64            `json:"actualLoops"`
	ExclusiveTime       float64            `json:"exclusiveTime"`     // ms (actualTotalTime minus children)
	PercentageOfTotal   float64            `json:"percentageOfTotal"` // percentage of query duration
	SharedHitBlocks     int64              `json:"sharedHitBlocks"`
	SharedReadBlocks    int64              `json:"sharedReadBlocks"`
	SharedDirtiedBlocks int64              `json:"sharedDirtiedBlocks"`
	SharedWrittenBlocks int64              `json:"sharedWrittenBlocks"`
	TempReadBlocks      int64              `json:"tempReadBlocks"`
	TempWrittenBlocks   int64              `json:"tempWrittenBlocks"`
	SortMethod          string             `json:"sortMethod,omitempty"`
	SortSpaceUsed       int64              `json:"sortSpaceUsed,omitempty"`
	SortSpaceType       string             `json:"sortSpaceType,omitempty"`
	Filter              string             `json:"filter,omitempty"`
	IndexCond           string             `json:"indexCond,omitempty"`
	HashCond            string             `json:"hashCond,omitempty"`
	RecheckCond         string             `json:"recheckCond,omitempty"`
	RowsRemovedByFilter int64              `json:"rowsRemovedByFilter,omitempty"`
	Plans               []*ExplainPlanNode `json:"plans,omitempty"`
}

// Bottleneck defines a flagged performance problem detected in an execution plan.
type Bottleneck struct {
	Type           string  `json:"type"` // HEAVY_SEQ_SCAN, DISK_SPILL, TIME_SINK, STALE_STATS, CACHE_MISS, FILTER_REMOVAL
	Severity       string  `json:"severity"` // CRITICAL, WARNING, INFO
	NodeID         string  `json:"nodeId"`
	NodeType       string  `json:"nodeType"`
	Relation       string  `json:"relation,omitempty"`
	Message        string  `json:"message"`
	Recommendation string  `json:"recommendation"`
	ImpactScore    float64 `json:"impactScore"` // 0.0 to 100.0
}

// ExplainAnalysisResult is the structured result returned after parsing and analyzing a plan.
type ExplainAnalysisResult struct {
	RootNode           *ExplainPlanNode   `json:"rootNode"`
	TotalExecutionTime float64            `json:"totalExecutionTime"` // ms
	PlanningTime       float64            `json:"planningTime"`       // ms
	TotalBuffersHit    int64              `json:"totalBuffersHit"`
	TotalBuffersRead   int64              `json:"totalBuffersRead"`
	CacheHitRatio      float64            `json:"cacheHitRatio"`
	Bottlenecks        []Bottleneck       `json:"bottlenecks"`
	FlattenedNodes     []*ExplainPlanNode `json:"flattenedNodes"`
}

// IndexRecommendation holds synthesized CREATE INDEX CONCURRENTLY suggestions.
type IndexRecommendation struct {
	ID                     string   `json:"id"`
	TableName              string   `json:"tableName"`
	IndexName              string   `json:"indexName"`
	Columns                []string `json:"columns"`
	Predicate              string   `json:"predicate,omitempty"`
	Reason                 string   `json:"reason"`
	DDL                    string   `json:"ddl"`
	EstimatedScanReduction float64  `json:"estimatedScanReduction"` // percentage
	EstimatedSizeBytes     int64    `json:"estimatedSizeBytes"`     // bytes
	EstimatedSizePretty    string   `json:"estimatedSizePretty"`    // e.g. "48.2 MB"
	WritePenalty           string   `json:"writePenalty"`           // LOW, MODERATE, SIGNIFICANT
	SeqScanCount           int64    `json:"seqScanCount"`
	SeqTuplesRead          int64    `json:"seqTuplesRead"`
	Priority               string   `json:"priority"` // HIGH, MEDIUM, LOW
}

// UnusedIndex represents an index with zero or negligible scans or redundancy.
type UnusedIndex struct {
	SchemaName      string   `json:"schemaName"`
	TableName       string   `json:"tableName"`
	IndexName       string   `json:"indexName"`
	IndexSizeBytes  int64    `json:"indexSizeBytes"`
	IndexSizePretty string   `json:"indexSizePretty"`
	IndexScans      int64    `json:"indexScans"`
	TuplesRead      int64    `json:"tuplesRead"`
	TuplesFetched   int64    `json:"tuplesFetched"`
	IsRedundant     bool     `json:"isRedundant"`
	RedundantWith   string   `json:"redundantWith,omitempty"`
	Columns         []string `json:"columns"`
	DropDDL         string   `json:"dropDdl"`
	Reason          string   `json:"reason"`
	Priority        string   `json:"priority"` // HIGH, MEDIUM, LOW
}

// LiveActivityItem describes an active or blocked database process.
type LiveActivityItem struct {
	PID           int64    `json:"pid"`
	User          string   `json:"user"`
	Database      string   `json:"database"`
	State         string   `json:"state"` // active, idle in transaction, etc.
	DurationMs    float64  `json:"durationMs"`
	WaitEventType string   `json:"waitEventType,omitempty"` // Lock, IO, Client, etc.
	WaitEvent     string   `json:"waitEvent,omitempty"`
	Query         string   `json:"query"`
	BlockingPID   int64    `json:"blockingPid,omitempty"`
	BlockedPIDs   []int64  `json:"blockedPids,omitempty"`
}

// LockContentionTree represents the blocked query topology.
type LockContentionTree struct {
	RootBlockers        []LiveActivityItem `json:"rootBlockers"`
	ActiveProcesses     []LiveActivityItem `json:"activeProcesses"`
	TotalBlockedQueries int                `json:"totalBlockedQueries"`
	ActiveLocksCount    int                `json:"activeLocksCount"`
	HasContention       bool               `json:"hasContention"`
}

// PlanNodeDiff describes a difference between two execution plan nodes.
type PlanNodeDiff struct {
	Path          string  `json:"path"`
	BaselineType  string  `json:"baselineType"`
	CurrentType   string  `json:"currentType"`
	BaselineTime  float64 `json:"baselineTime"`
	CurrentTime   float64 `json:"currentTime"`
	TimeDeltaMs   float64 `json:"timeDeltaMs"`
	ChangeSummary string  `json:"changeSummary"`
}

// PlanDiffResult models the delta between two execution plans.
type PlanDiffResult struct {
	BaselineExecutionTime float64        `json:"baselineExecutionTime"` // ms
	CurrentExecutionTime  float64        `json:"currentExecutionTime"`  // ms
	TimeDeltaMs           float64        `json:"timeDeltaMs"`           // ms (negative = faster)
	TimeDeltaPct          float64        `json:"timeDeltaPct"`          // percentage
	BaselineBuffersHit    int64          `json:"baselineBuffersHit"`
	CurrentBuffersHit     int64          `json:"currentBuffersHit"`
	BaselineBuffersRead   int64          `json:"baselineBuffersRead"`
	CurrentBuffersRead    int64          `json:"currentBuffersRead"`
	NodeChanges           []PlanNodeDiff `json:"nodeChanges"`
	Verdict               string         `json:"verdict"` // IMPROVED, REGRESSED, NEUTRAL
	Summary               string         `json:"summary"`
}

// HypoSimulationResult stores the simulated cost impact of an index via HypoPG.
type HypoSimulationResult struct {
	Supported          bool    `json:"supported"`
	ExtensionInstalled bool    `json:"extensionInstalled"`
	OriginalCost       float64 `json:"originalCost"`
	SimulatedCost      float64 `json:"simulatedCost"`
	CostReductionPct   float64 `json:"costReductionPct"`
	OriginalPlanType   string  `json:"originalPlanType"`
	SimulatedPlanType  string  `json:"simulatedPlanType"`
	IndexName          string  `json:"indexName"`
	Columns            string  `json:"columns"`
	Message            string  `json:"message"`
}

// BloatInfo represents estimated table and index disk bloat.
type BloatInfo struct {
	TableName      string  `json:"tableName"`
	SizeBytes      int64   `json:"sizeBytes"`
	BloatBytes     int64   `json:"bloatBytes"`
	BloatRatio     float64 `json:"bloatRatio"` // percentage
	LiveTuples     int64   `json:"liveTuples"`
	DeadTuples     int64   `json:"deadTuples"`
	DeadTupleRatio float64 `json:"deadTupleRatio"`
	SizePretty     string  `json:"sizePretty"`
	BloatPretty    string  `json:"bloatPretty"`
}

// AutovacuumWorker models an active worker in pg_stat_progress_vacuum.
type AutovacuumWorker struct {
	PID                int64   `json:"pid"`
	Database           string  `json:"database"`
	Table              string  `json:"table"`
	Phase              string  `json:"phase"`
	HeapBlksTotal      int64   `json:"heapBlksTotal"`
	HeapBlksScanned    int64   `json:"heapBlksScanned"`
	HeapBlksVacuumed   int64   `json:"heapBlksVacuumed"`
	IndexVacuumCount   int64   `json:"indexVacuumCount"`
	ProgressPercentage float64 `json:"progressPercentage"`
}

// WraparoundInfo tracks transaction ID consumption against freeze limits.
type WraparoundInfo struct {
	DatabaseName    string  `json:"databaseName"`
	CurrentTxidAge  int64   `json:"currentTxidAge"`
	MaxTxidAge      int64   `json:"maxTxidAge"`
	PercentConsumed float64 `json:"percentConsumed"`
	RemainingTxids  int64   `json:"remainingTxids"`
	Status          string  `json:"status"` // HEALTHY, WARNING, CRITICAL
}

// SystemOverview aggregates cockpit health signals.
type SystemOverview struct {
	Connected               bool    `json:"connected"`
	DemoMode                bool    `json:"demoMode"`
	DatabaseVersion         string  `json:"databaseVersion"`
	DatabaseName            string  `json:"databaseName"`
	ActiveConnections       int     `json:"activeConnections"`
	MaxConnections          int     `json:"maxConnections"`
	BufferCacheHitRatio     float64 `json:"bufferCacheHitRatio"`
	TotalQueriesTracked     int     `json:"totalQueriesTracked"`
	UnderIndexedTablesCount int     `json:"underIndexedTablesCount"`
	UnusedIndexesCount      int     `json:"unusedIndexesCount"`
	ActiveLocksCount        int     `json:"activeLocksCount"`
	BloatWarningCount       int     `json:"bloatWarningCount"`
	WraparoundSafetyPercent float64 `json:"wraparoundSafetyPercent"`
	ActiveAutovacuums       int     `json:"activeAutovacuums"`
}

// SQLAntiPattern defines common sub-optimal query construction and remediation.
type SQLAntiPattern struct {
	Type          string `json:"type"`          // FUNCTION_ON_INDEX_COLUMN, LEADING_WILDCARD, DEEP_OFFSET, SELECT_STAR
	Severity      string `json:"severity"`      // CRITICAL, WARNING, SUGGESTION
	Snippet       string `json:"snippet"`       // offending SQL fragment
	Message       string `json:"message"`       // explanation of performance penalty
	RewriteAdvice string `json:"rewriteAdvice"` // concrete rewritten SQL advice
	SuggestedDDL  string `json:"suggestedDdl,omitempty"` // CREATE INDEX CONCURRENTLY or extension DDL
}

// ConfigSetting represents a tuned or active PostgreSQL server parameter.
type ConfigSetting struct {
	Name             string `json:"name"`
	CurrentValue     string `json:"currentValue"`
	RecommendedValue string `json:"recommendedValue"`
	Unit             string `json:"unit"`
	Category         string `json:"category"` // Memory, Query Planner, WAL, Concurrency
	RestartRequired  bool   `json:"restartRequired"`
	Reason           string `json:"reason"`
	Status           string `json:"status"` // OPTIMAL, ATTENTION, CRITICAL
	AlterDDL         string `json:"alterDdl"`
}

// TuningProfile defines the hardware capacity and workload shape.
type TuningProfile struct {
	TotalRAMGB   float64 `json:"totalRamGb"`
	VCPUs        int     `json:"vcpus"`
	DiskType     string  `json:"diskType"`     // NVMe, SSD, HDD
	WorkloadType string  `json:"workloadType"` // WEB, OLTP, DW, MIXED
}

// TuningReport aggregates recommended configuration settings against current database state.
type TuningReport struct {
	Profile        TuningProfile   `json:"profile"`
	Settings       []ConfigSetting `json:"settings"`
	OptimalCount   int             `json:"optimalCount"`
	AttentionCount int             `json:"attentionCount"`
	AlterScript    string          `json:"alterScript"`
}

// UnindexedForeignKey identifies child foreign key constraints lacking supporting B-Tree indexes.
type UnindexedForeignKey struct {
	ConstraintName string   `json:"constraintName"`
	ChildTable     string   `json:"childTable"`
	ParentTable    string   `json:"parentTable"`
	Columns        []string `json:"columns"`
	SuggestedDDL   string   `json:"suggestedDdl"`
	Severity       string   `json:"severity"` // CRITICAL, WARNING
	Reason         string   `json:"reason"`
}

// TableMissingPK identifies relations operating without a declared Primary Key.
type TableMissingPK struct {
	SchemaName      string `json:"schemaName"`
	TableName       string `json:"tableName"`
	EstimatedRows   int64  `json:"estimatedRows"`
	TotalSizePretty string `json:"totalSizePretty"`
	SuggestedDDL    string `json:"suggestedDdl"`
}

// ToastRelationInfo details oversized TOAST storage ratios relative to heap tables.
type ToastRelationInfo struct {
	TableName       string  `json:"tableName"`
	ToastTableName  string  `json:"toastTableName"`
	MainSizeBytes   int64   `json:"mainSizeBytes"`
	ToastSizeBytes  int64   `json:"toastSizeBytes"`
	MainSizePretty  string  `json:"mainSizePretty"`
	ToastSizePretty string  `json:"toastSizePretty"`
	ToastRatio      float64 `json:"toastRatio"` // percentage (toast / total)
	Recommendation  string  `json:"recommendation"`
}

// SchemaHealthReport combines missing foreign key indexes, missing primary keys, and TOAST bloat.
type SchemaHealthReport struct {
	UnindexedFKs   []UnindexedForeignKey `json:"unindexedFks"`
	MissingPKs     []TableMissingPK      `json:"missingPks"`
	ToastRelations []ToastRelationInfo   `json:"toastRelations"`
	TotalIssues    int                   `json:"totalIssues"`
	HealthScore    int                   `json:"healthScore"` // 0 to 100
}

// ConnectionPoolHealth captures connection state saturation and sizing efficiency.
type ConnectionPoolHealth struct {
	TotalConnections    int            `json:"totalConnections"`
	MaxConnections      int            `json:"maxConnections"`
	UtilizationPercent  float64        `json:"utilizationPercent"`
	ActiveCount         int            `json:"activeCount"`
	IdleCount           int            `json:"idleCount"`
	IdleInTxCount       int            `json:"idleInTxCount"`
	MaxIdleInTxDuration float64        `json:"maxIdleInTxDuration"` // seconds
	StateBreakdown      map[string]int `json:"stateBreakdown"`
	OptimalPoolSize     int            `json:"optimalPoolSize"`
	Status              string         `json:"status"` // HEALTHY, WARNING, CRITICAL
	Warnings            []string       `json:"warnings"`
	PgBouncerAdvice     string         `json:"pgBouncerAdvice"`
}

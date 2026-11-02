export interface QueryStat {
  queryId: string;
  query: string;
  calls: number;
  totalExecTime: number; // ms
  meanExecTime: number;  // ms
  minExecTime: number;
  maxExecTime: number;
  stddevExecTime: number;
  rows: number;
  rowsPerCall: number;
  sharedBlksHit: number;
  sharedBlksRead: number;
  sharedBlksDirtied: number;
  sharedBlksWritten: number;
  tempBlksRead: number;
  tempBlksWritten: number;
  cacheHitRatio: number; // percentage
}

export interface TableStat {
  schemaName: string;
  tableName: string;
  seqScan: number;
  seqTupRead: number;
  idxScan: number;
  idxTupFetch: number;
  liveTuples: number;
  deadTuples: number;
  deadTupleRatio: number;
  tableSizeBytes: number;
  indexSizeBytes: number;
  tableSizePretty: string;
  indexSizePretty: string;
  lastVacuum: string | null;
  lastAutovacuum: string | null;
}

export interface TableCacheStat {
  tableName: string;
  heapBlksHit: number;
  heapBlksRead: number;
  heapHitRatio: number;
  indexBlksHit: number;
  indexBlksRead: number;
  indexHitRatio: number;
}

export interface CacheHitStats {
  globalHeapHitRatio: number;
  globalIndexHitRatio: number;
  overallHitRatio: number;
  totalHeapHit: number;
  totalHeapRead: number;
  totalIndexHit: number;
  totalIndexRead: number;
  tablesBelowThreshold: TableCacheStat[];
}

export interface ExplainPlanNode {
  nodeId: string;
  nodeType: string;
  relationName?: string;
  schema?: string;
  alias?: string;
  indexName?: string;
  startupCost: number;
  totalCost: number;
  planRows: number;
  planWidth: number;
  actualStartupTime: number;
  actualTotalTime: number;
  actualRows: number;
  actualLoops: number;
  exclusiveTime: number;
  percentageOfTotal: number;
  sharedHitBlocks: number;
  sharedReadBlocks: number;
  sharedDirtiedBlocks: number;
  sharedWrittenBlocks: number;
  tempReadBlocks: number;
  tempWrittenBlocks: number;
  sortMethod?: string;
  sortSpaceUsed?: number;
  sortSpaceType?: string;
  filter?: string;
  indexCond?: string;
  hashCond?: string;
  recheckCond?: string;
  rowsRemovedByFilter?: number;
  plans?: ExplainPlanNode[];
}

export interface Bottleneck {
  type: string;
  severity: 'CRITICAL' | 'WARNING' | 'INFO';
  nodeId: string;
  nodeType: string;
  relation?: string;
  message: string;
  recommendation: string;
  impactScore: number;
}

export interface ExplainAnalysisResult {
  rootNode: ExplainPlanNode;
  totalExecutionTime: number;
  planningTime: number;
  totalBuffersHit: number;
  totalBuffersRead: number;
  cacheHitRatio: number;
  bottlenecks: Bottleneck[];
  flattenedNodes: ExplainPlanNode[];
}

export interface IndexRecommendation {
  id: string;
  tableName: string;
  indexName: string;
  columns: string[];
  predicate?: string;
  reason: string;
  ddl: string;
  estimatedScanReduction: number;
  seqScanCount: number;
  seqTuplesRead: number;
  priority: 'HIGH' | 'MEDIUM' | 'LOW';
}

export interface BloatInfo {
  tableName: string;
  sizeBytes: number;
  bloatBytes: number;
  bloatRatio: number;
  liveTuples: number;
  deadTuples: number;
  deadTupleRatio: number;
  sizePretty: string;
  bloatPretty: string;
}

export interface AutovacuumWorker {
  pid: number;
  database: string;
  table: string;
  phase: string;
  heapBlksTotal: number;
  heapBlksScanned: number;
  heapBlksVacuumed: number;
  indexVacuumCount: number;
  progressPercentage: number;
}

export interface WraparoundInfo {
  databaseName: string;
  currentTxidAge: number;
  maxTxidAge: number;
  percentConsumed: number;
  remainingTxids: number;
  status: 'HEALTHY' | 'WARNING' | 'CRITICAL';
}

export interface SystemOverview {
  connected: boolean;
  demoMode: boolean;
  databaseVersion: string;
  databaseName: string;
  activeConnections: number;
  maxConnections: number;
  bufferCacheHitRatio: number;
  totalQueriesTracked: number;
  underIndexedTablesCount: number;
  bloatWarningCount: number;
  wraparoundSafetyPercent: number;
  activeAutovacuums: number;
}

# IndexLens: Master Implementation Plan

> **Non-Invasive PostgreSQL Slow Query Explainer, Autovacuum Watchdog & Missing Index Synthesizer**
>
> High-performance database optimization cockpit and query diagnostics engine for PostgreSQL. Reads non-invasively from `pg_stat_statements`, visualizes `EXPLAIN (ANALYZE, BUFFERS)` plans as interactive execution flamegraphs, synthesizes non-blocking `CREATE INDEX CONCURRENTLY` suggestions, and guards against autovacuum lag and table bloat.

---

## 1. Project Overview & Architecture

PostgreSQL powers modern transactional applications. As data sets expand, engineering teams face key challenges:
1. **Cryptic Execution Plans**: Developers run `EXPLAIN (ANALYZE, BUFFERS)` and receive 80-line nested textual trees. Isolating whether latency was caused by sequential table scans, buffer cache misses, or disk spills is slow and error-prone.
2. **Missing Indexes & Locking Risks**: Queries filtering on multiple columns scan millions of rows because only single-column indexes exist. Standard `CREATE INDEX` locks tables against writes, risking downtime.
3. **The Autovacuum & Bloat Trap**: PostgreSQL uses Multi-Version Concurrency Control (MVCC). If autovacuum settings lag behind write velocity, tables and indexes accumulate dead tuples, degrading cache hit ratios and eventually risking catastrophic Transaction ID (TxID) Wraparound outages.
4. **Cloud SaaS Expense and Data Privacy**: Cloud database monitoring products charge hundreds of dollars per instance and require exfiltrating query text and metadata outside corporate security boundaries.

IndexLens delivers a non-invasive, open-source, local-first performance cockpit:
- **Zero-Impact Polling**: Connects with read-only credentials to `pg_stat_statements`, `pg_stat_user_tables`, `pg_statio_user_tables`, and `pg_stat_progress_vacuum`.
- **Interactive EXPLAIN Flamegraph**: Transforms complex JSON execution plans into visual flamegraphs, color-coding high-cost nodes, buffer cache misses, and memory sorts.
- **Missing Index Synthesizer**: Correlates sequential scans with active query filter patterns, automatically generating non-blocking DDL:
  `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_created ON orders (user_id, created_at DESC);`.
- **Autovacuum Watchdog & Bloat Estimator**: Calculates dead tuple ratios, tracks autovacuum worker queue lag, and alerts before transaction ID wraparound thresholds are approached.
- **Dual Mode Operation**: Connects to live PostgreSQL databases or runs in autonomous demo mode with realistic production telemetry for offline evaluation.

```
+-----------------------------------------------------------------------------------+
|                                     INDEXLENS                                     |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|  Non-Invasive Metadata Poller   |                 |    Visual EXPLAIN Flamegraph  |
|  * pg_stat_statements Metrics   |                 |  * EXPLAIN (ANALYZE, BUFFERS) |
|  * pg_stat_user_tables Scans    |                 |  * Node Cost & Duration Map   |
|  * pg_statio Buffer Cache Hits  |                 |  * Disk Spill & Temp File Warn|
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |    Missing Index Synthesizer      |
                        |  * Seq Scan to Filter Correlation |
                        |  * Composite B-Tree Column Ranker |
                        |  * CREATE INDEX CONCURRENTLY DDL  |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Autovacuum & Bloat Watchdog     |
                        |  * Dead Tuple Ratio Calculation   |
                        |  * Table & Index Bloat Estimator  |
                        |  * TXID Wraparound Safety Horizon |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Interactive Web Studio & CLI    |
                        |  * Query Latency Cockpit          |
                        |  * Cache Hit Gauge (> 99%)        |
                        |  * 1-Click DDL Copy (Port 9920)   |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
indexlens/
├── cmd/
│   └── indexlens/
│       └── main.go               # Unified CLI: serve, explain, bloat, advice
├── internal/
│   ├── model/
│   │   ├── types.go              # Core data structs for queries, plans, tables, bloat
│   │   └── demo_data.go          # Realistic production simulation telemetry
│   ├── poller/
│   │   ├── client.go             # Read-only PostgreSQL connection pool (pgx/v5)
│   │   ├── stat_statements.go    # Query frequency, mean/total duration, buffer metrics
│   │   ├── stat_tables.go        # seq_scan vs idx_scan & live/dead tuples
│   │   └── statio.go             # Heap & index buffer cache hit ratio calculator
│   ├── explain/
│   │   ├── parser.go             # EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS) AST parser
│   │   ├── node_tree.go          # Recursive execution node hierarchy & flamegraph builder
│   │   └── bottlenecks.go        # Identifies slow nodes, disk spills, and high buffers
│   ├── synthesizer/
│   │   ├── index_advisor.go      # Correlates filter predicates with high-cost scans
│   │   ├── column_ranker.go      # Orders equality columns before range and sort keys
│   │   └── ddl_generator.go      # Generates safe CREATE INDEX CONCURRENTLY statements
│   ├── watchdog/
│   │   ├── bloat_estimator.go    # Estimates wasted disk space in tables and indexes
│   │   ├── autovacuum_queue.go   # Monitors active vacuum workers in pg_stat_progress_vacuum
│   │   └── wraparound_guard.go   # Calculates txid age vs autovacuum_freeze_max_age
│   └── server/
│       ├── api.go                # REST API handlers
│       ├── router.go             # HTTP multiplexer with CORS and logging
│       └── static.go             # Embedded React frontend assets (embed.FS)
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Header.tsx            # Navigation, connection status, quick stats
│   │   │   ├── SlowQueryTable.tsx    # Queries sorted by total time, calls, mean time
│   │   │   ├── ExplainVisualizer.tsx # Visual execution plan flamegraph & node inspector
│   │   │   ├── IndexAdvisorCard.tsx  # Recommended indexes with 1-click SQL copy
│   │   │   ├── BloatWatchdog.tsx     # Table dead tuples and wasted space chart
│   │   │   └── CacheHitGauge.tsx     # Buffer cache hit ratio gauge
│   │   ├── hooks/
│   │   │   └── useStats.ts           # Data fetching and polling hook
│   │   ├── types/
│   │   │   └── api.ts                # TypeScript definitions matching Go models
│   │   ├── App.tsx                   # Main dashboard layout
│   │   ├── index.css                 # Tailwind CSS styles
│   │   └── main.tsx
│   ├── package.json
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs/
│   ├── ARCHITECTURE.md
│   ├── POSTGRESQL_SETUP.md
│   ├── INDEX_SYNTHESIS_RULES.md
│   └── AUTOVACUUM_AND_BLOAT.md
├── go.mod
├── go.sum
├── LICENSE
├── CONTRIBUTING.md
└── README.md
```

---

## 3. Step-by-Step Implementation Roadmap

### Step 1: Go Module & Core Domain Models
- Initialize Go module `github.com/alexandrmotologa/indexlens`.
- Define unified domain models in `internal/model/types.go`:
  - `QueryStat`: Query ID, SQL text, calls, total time, mean time, min/max time, rows, shared blocks hit/read, local/temp blocks.
  - `TableStat`: Schema, table name, sequential scans, index scans, live tuples, dead tuples, dead ratio, last vacuum/autovacuum.
  - `CacheHitStats`: Global heap/index hit ratios, per-table cache metrics.
  - `ExplainPlanNode`: Recursive plan tree, node type, relation name, index name, cost, startup time, total time, actual rows, loops, shared/temp buffers, sort method, filter predicate.
  - `Bottleneck`: Severity, node ID, message, recommendation, metric.
  - `IndexRecommendation`: Table, suggested index name, columns, reason, estimated impact, non-blocking DDL.
  - `BloatInfo`: Table name, size bytes, bloat bytes, bloat ratio, dead tuples.
  - `WraparoundInfo`: Database name, current txid age, max age limit, safety percentage, danger status.
  - `AutovacuumWorker`: PID, table, phase, heap blocks scanned, total blocks, progress percentage.
- Implement `internal/model/demo_data.go` providing realistic PostgreSQL statistics for demo/offline execution.

### Step 2: Non-Invasive Database Poller (`internal/poller`)
- Implement `client.go` with `jackc/pgx/v5/pgxpool`:
  - Read-only connection establishment.
  - Feature detection: verify `pg_stat_statements` is installed; handle graceful degradation if missing.
- Implement `stat_statements.go`:
  - Query `pg_stat_statements` sorting by total time or mean time.
  - Calculate per-query cache hit ratio: `shared_blks_hit / (shared_blks_hit + shared_blks_read) * 100`.
- Implement `stat_tables.go`:
  - Query `pg_stat_user_tables` to collect sequential scan vs index scan counts and live/dead tuple counts.
- Implement `statio.go`:
  - Query `pg_statio_user_tables` and `pg_statio_user_indexes` to compute buffer cache hit rates.

### Step 3: EXPLAIN Parser & Visual Flamegraph Engine (`internal/explain`)
- Implement `parser.go`:
  - Parse PostgreSQL `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` JSON structures.
  - Build recursive `PlanNode` hierarchy with inclusive time and computed exclusive time.
- Implement `node_tree.go`:
  - Calculate width percentages for flamegraph rendering.
  - Extract buffer read vs hit proportions per node.
- Implement `bottlenecks.go`:
  - Detect high runtime sinks (> 40% query duration).
  - Detect disk spills (`external merge Disk`, `Workfile Used`).
  - Detect heavy sequential scans on large tables.
  - Detect stale statistics: discrepancy between `Plan Rows` and `Actual Rows` > 10x.

### Step 4: Missing Index Synthesizer (`internal/synthesizer`)
- Implement `column_ranker.go`:
  - Order equality columns (`=`, `IS NULL`) first.
  - Order range/inequality columns (`<`, `>`, `BETWEEN`, `IN`, `LIKE`) second.
  - Order sort keys (`ORDER BY`) last.
- Implement `index_advisor.go`:
  - Correlate sequential scan tables with filter conditions from execution plans and slow queries.
  - Detect partial index opportunities (`WHERE deleted_at IS NULL` or `WHERE status = 'pending'`).
- Implement `ddl_generator.go`:
  - Generate production-safe `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_... ON ... (...)`.
  - Provide estimated impact notes.

### Step 5: Autovacuum Watchdog & Bloat Estimator (`internal/watchdog`)
- Implement `bloat_estimator.go`:
  - Compute dead tuple ratios and estimate physical page bloat.
- Implement `autovacuum_queue.go`:
  - Query `pg_stat_progress_vacuum` for active workers, current phase, and completion progress.
- Implement `wraparound_guard.go`:
  - Query `age(datfrozenxid)` and compute headroom against `autovacuum_freeze_max_age`.

### Step 6: Web Studio Backend & REST API (`internal/server`)
- Implement REST API endpoints:
  - `GET /api/health`: Health and connection state.
  - `GET /api/overview`: Global metrics, cache hit ratio, active connections, bloat stats.
  - `GET /api/queries`: Slow queries with search and sort.
  - `POST /api/explain`: Ingests JSON plan and returns flamegraph nodes and bottleneck diagnostics.
  - `GET /api/tables`: Table statistics and dead tuple ratios.
  - `GET /api/indexes/advisor`: Synthesized index suggestions with copyable DDL.
  - `GET /api/autovacuum`: Active workers and TxID wraparound metrics.
- Implement static asset serving with embedded files (`embed.FS`).

### Step 7: React 19 Web Studio (`web/`)
- Initialize Vite React 19 + TypeScript + Tailwind CSS application.
- Build components:
  - `Header`: Connection status, cache hit ratio badge, wraparound safety badge, demo mode toggle.
  - `SlowQueryTable`: Sortable table with search, mean latency, total time, calls, and buffer hit breakdown.
  - `ExplainVisualizer`: Interactive flamegraph view with zoom, node inspector, and bottleneck alert banners.
  - `IndexAdvisorCard`: Recommended indexes with 1-click SQL copy, explanation, and impact estimation.
  - `BloatWatchdog`: Dead tuples bar chart, table bloat table, and running autovacuum progress bar.
  - `CacheHitGauge`: Radial / bar gauge showing heap and index cache hit rates with 99% threshold alert.
- Compile and bundle frontend assets into `internal/server/dist/`.

### Step 8: CLI Commands & End-to-End Verification (`cmd/indexlens`)
- Implement unified CLI:
  - `indexlens serve [--dsn ...] [--port 9920] [--demo]`
  - `indexlens explain --file plan.json`
  - `indexlens bloat [--dsn ...]`
  - `indexlens advice [--dsn ...]`
- Write comprehensive unit tests for all packages:
  - `internal/poller`
  - `internal/explain`
  - `internal/synthesizer`
  - `internal/watchdog`
  - `internal/server`
- Verify binary builds cleanly and all tests pass with 100% green status.

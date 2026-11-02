# IndexLens: Master Implementation Plan

> **Non-Invasive PostgreSQL Slow Query Explainer, Autovacuum Watchdog & Missing Index Synthesizer**
>
> High-performance database optimization cockpit and query diagnostics engine for PostgreSQL. Reads non-invasively from `pg_stat_statements`, visualizes `EXPLAIN (ANALYZE, BUFFERS)` plans as interactive execution flamegraphs, synthesizes non-blocking `CREATE INDEX CONCURRENTLY` suggestions, guards against autovacuum lag, table bloat, and lock contention, and provides regression plan diffs and unused index detection.

---

## 1. Project Overview & Architecture

PostgreSQL powers modern transactional applications. As data sets expand, engineering teams face key challenges:
1. **Cryptic Execution Plans**: Developers run `EXPLAIN (ANALYZE, BUFFERS)` and receive 80-line nested textual trees. Isolating whether latency was caused by sequential table scans, buffer cache misses, or disk spills is slow and error-prone.
2. **Missing Indexes & Write Penalties**: Queries filtering on multiple columns scan millions of rows because only single-column indexes exist, while unused or redundant indexes penalize every write and waste gigabytes of `shared_buffers`.
3. **The Autovacuum & Bloat Trap**: If autovacuum settings lag behind write velocity, tables and indexes accumulate dead tuples, degrading cache hit ratios and eventually risking catastrophic Transaction ID (TxID) Wraparound outages.
4. **Lock Contention Blindspots**: Spikes in latency are often caused by uncommitted locks or `idle in transaction` queries holding locks on critical relations.

IndexLens delivers a non-invasive, open-source, local-first performance cockpit:
- **Zero-Impact Polling**: Connects with read-only credentials to `pg_stat_statements`, `pg_stat_user_tables`, `pg_statio_user_tables`, and `pg_stat_activity`.
- **Interactive EXPLAIN Flamegraph**: Transforms complex JSON execution plans into visual flamegraphs, color-coding high-cost nodes, buffer cache misses, and memory sorts.
- **Missing Index Synthesizer**: Correlates sequential scans with active query filter patterns, automatically generating non-blocking DDL with index size and write-penalty projections.
- **Unused & Redundant Index Cleaner**: Identifies zero-scan and redundant prefix indexes with non-blocking `DROP INDEX CONCURRENTLY` DDL.
- **Plan Diff & Regression Analyzer**: Compares execution plans side-by-side to verify speedups or catch query plan regressions.
- **Live Lock Contention Graph**: Visualizes blocking transaction dependencies and connection wait events.
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
|  * pg_stat_activity Lock Graph  |                 |  * Plan Diff Comparison View  |
|  * pg_statio Buffer Cache Hits  |                 |  * Disk Spill & Temp File Warn|
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |    Missing & Unused Index Engine  |
                        |  * Seq Scan to Filter Correlation |
                        |  * Composite B-Tree Column Ranker |
                        |  * CREATE INDEX CONCURRENTLY DDL  |
                        |  * Unused & Redundant Index DROP  |
                        |  * In-Memory HypoPG Verification  |
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
                        |  * Real-Time Lock Contention Tree |
                        |  * CI/CD Quality Gate Check       |
                        |  * 1-Click Markdown Report Export |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
indexlens/
├── cmd/
│   └── indexlens/
│       └── main.go               # Unified CLI: serve, explain, bloat, advice, unused, locks, check, report
├── internal/
│   ├── model/
│   │   ├── types.go              # Core data structs for queries, plans, tables, bloat, locks, unused
│   │   └── demo_data.go          # Realistic production simulation telemetry
│   ├── poller/
│   │   ├── client.go             # Read-only PostgreSQL connection pool (pgx/v5)
│   │   ├── stat_statements.go    # Query frequency, mean/total duration, buffer metrics
│   │   ├── stat_tables.go        # seq_scan vs idx_scan & live/dead tuples
│   │   ├── statio.go             # Heap & index buffer cache hit ratio calculator
│   │   ├── unused_indexes.go     # Zero-scan and redundant prefix index detector
│   │   └── locks.go              # Real-time lock contention and blocking PID tree
│   ├── explain/
│   │   ├── parser.go             # EXPLAIN (FORMAT JSON, ANALYZE, BUFFERS) AST parser
│   │   ├── node_tree.go          # Recursive execution node hierarchy & flamegraph builder
│   │   ├── bottlenecks.go        # Identifies slow nodes, disk spills, and high buffers
│   │   └── diff.go               # Plan diff engine comparing baseline vs current
│   ├── synthesizer/
│   │   ├── index_advisor.go      # Correlates filter predicates with high-cost scans
│   │   ├── column_ranker.go      # Orders equality columns before range and sort keys
│   │   ├── ddl_generator.go      # Generates safe CREATE INDEX CONCURRENTLY statements
│   │   └── hypo.go               # B-Tree index size estimator & HypoPG simulator
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
│   │   │   ├── Header.tsx            # Navigation, connection status, quick stats, report export
│   │   │   ├── SlowQueryTable.tsx    # Queries sorted by total time, calls, mean time
│   │   │   ├── ExplainVisualizer.tsx # Visual execution plan flamegraph & node inspector
│   │   │   ├── PlanDiffViewer.tsx    # Side-by-side plan diff with timing and node transformation deltas
│   │   │   ├── IndexAdvisorCard.tsx  # Recommended indexes + Unused/Redundant cleanup switcher
│   │   │   ├── UnusedIndexList.tsx   # Unused & duplicate indexes list with safe DROP DDL
│   │   │   ├── LockContentionView.tsx# Real-time lock contention tree & wait events
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

## 3. Implementation Status (All Phases Completed)

- [x] **Phase 1: Go Module & Core Domain Models**: Initialized Go module and structured domain models.
- [x] **Phase 2: Non-Invasive Database Poller**: Implemented `pg_stat_statements`, `pg_stat_user_tables`, `pg_statio_user_tables`.
- [x] **Phase 3: EXPLAIN Parser & Visual Flamegraph Engine**: AST tree generator, exclusive duration calculator, and bottleneck detector.
- [x] **Phase 4: Missing Index Synthesizer**: B-Tree column ranker (Equality $\to$ Range $\to$ Sort), partial index detection, and `CREATE INDEX CONCURRENTLY` DDL generator.
- [x] **Phase 5: Autovacuum Watchdog & Bloat Estimator**: Table bloat calculations, worker queue tracking, and TxID wraparound safety horizon.
- [x] **Phase 6: Web Studio Backend & REST API**: HTTP endpoints with embedded static asset distribution (`embed.FS`).
- [x] **Phase 7: React 19 Web Studio**: Interactive Cockpit, Flamegraph visualizer, Index cards, Bloat gauges.
- [x] **Phase 8: Unused & Redundant Index Cleaner**: Zero-scan detection and prefix index deduplication (`DROP INDEX CONCURRENTLY`).
- [x] **Phase 9: Live Lock Contention Tracker**: `pg_stat_activity` and `pg_locks` real-time blocker hierarchy.
- [x] **Phase 10: Plan Diff & Regression Engine**: Baseline vs Current plan comparison with structural and timing deltas.
- [x] **Phase 11: HypoPG In-Memory Verification & Index Sizing**: B-Tree byte sizing formula and HypoPG simulation.
- [x] **Phase 12: CI/CD Quality Gate & Report Exporter**: CLI `check` command with exit codes and Markdown report generation.

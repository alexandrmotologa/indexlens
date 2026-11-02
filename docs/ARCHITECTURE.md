# IndexLens Architecture

IndexLens is designed as a single-binary, local-first database performance monitor and query optimizer for PostgreSQL.

```
                      +------------------------------------------+
                      |         PostgreSQL Database Engine       |
                      |   pg_stat_statements  pg_statio_tables   |
                      |   pg_stat_user_tables pg_stat_activity   |
                      +--------------------+---------------------+
                                           | Read-Only Polling
                                           v
+----------------------------------------------------------------------------------+
|                                INDEXLENS ENGINE                                  |
|                                                                                  |
|   +-------------------+    +--------------------+    +-----------------------+   |
|   |  poller/client    |    |   explain/parser   |    | synthesizer/advisor   |   |
|   |  pgx connection   |    |   AST JSON parser  |    | Filter extraction     |   |
|   |  Stat statement   |    |   Node hierarchy   |    | B-Tree column ranker  |   |
|   |  Statio hit ratio |    |   Bottleneck rules |    | CONCURRENT DDL build  |   |
|   +---------+---------+    +---------+----------+    +-----------+-----------+   |
|             |                        |                           |               |
|             +------------------------+---------------------------+               |
|                                      |                                           |
|                           +----------v----------+                                |
|                           |   watchdog/bloat    |                                |
|                           |   Tuple dead ratio  |                                |
|                           |   Progress vacuum   |                                |
|                           |   Wraparound guard  |                                |
|                           +----------+----------+                                |
|                                      |                                           |
|                           +----------v----------+                                |
|                           |   server/api        |                                |
|                           |   REST endpoints    |                                |
|                           |   Embedded static   |                                |
|                           +----------+----------+                                |
+--------------------------------------|-------------------------------------------+
                                       | HTTP :9920
                                       v
                     +-----------------------------------+
                     |       React 19 Web Studio         |
                     |  - Slow Query Cockpit             |
                     |  - Interactive Explain Flamegraph |
                     |  - Missing Index DDL Cards        |
                     |  - Table Bloat & Autovacuum Gauges|
                     +-----------------------------------+
```

## Core Modules

### 1. Poller (`internal/poller`)
Manages read-only pooling to PostgreSQL via `github.com/jackc/pgx/v5`.
- **Query Latencies**: Gathers statistics from `pg_stat_statements` with aggregated metrics: calls, total time, mean time, rows, buffer hit ratios.
- **Cache Hit Telemetry**: Calculates database-level and table-level buffer cache hit rates:
  $$\text{HitRatio} = \frac{\text{heap\_blks\_hit} \times 100}{\text{heap\_blks\_hit} + \text{heap\_blks\_read}}$$
- **Table Activity**: Tracks `seq_scan` vs `idx_scan` frequencies to detect under-indexed workloads.

### 2. Explain Visualizer (`internal/explain`)
Parses execution plans output by PostgreSQL's `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` command:
- Transforms nested plan nodes into a recursive tree structure suitable for flamegraph rendering.
- Computes inclusive and exclusive execution times for each operation.
- Identifies critical bottlenecks:
  - High duration nodes (> 40% parent runtime).
  - External disk spills (`Sort Method: external merge Disk`).
  - Large sequential table scans on multi-thousand row relations.
  - Large discrepancies between planner estimated rows and actual execution rows.

### 3. Missing Index Synthesizer (`internal/synthesizer`)
Correlates identified sequential scans with active query filter patterns:
- Parses `WHERE`, `JOIN`, and `ORDER BY` predicates.
- Enforces optimal B-Tree column ordering:
  1. Equality operators (`=`, `IS NULL`).
  2. Range and inequality operators (`<`, `>`, `<=`, `>=`, `BETWEEN`, `LIKE`).
  3. Sort key columns (`ORDER BY`).
- Emits non-blocking SQL: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_... ON ... (...)`.

### 4. Autovacuum & Bloat Watchdog (`internal/watchdog`)
Maintains database health invariants:
- **Dead Tuple Ratio**: Identifies tables with elevated dead tuple percentages.
- **Table Bloat**: Estimates wasted disk space using tuple header width and page fill factor calculations.
- **Autovacuum Worker Tracking**: Reads `pg_stat_progress_vacuum` to display live progress and current vacuum phases.
- **Transaction ID Wraparound**: Monitors `age(datfrozenxid)` against `autovacuum_freeze_max_age` (200 million threshold).

### 5. Hardware & Configuration Tuner (`internal/poller/config_tuner.go`)
Calculates optimal system configuration based on RAM, vCPU count, and storage characteristics:
- Calculates `shared_buffers` (25% of RAM, capped at 16GB), `effective_cache_size` (75% of RAM), `work_mem` (connection-scaled), and `maintenance_work_mem`.
- Sets `random_page_cost` and `effective_io_concurrency` for NVMe/SSD vs HDD storage media.
- Emits reproducible, idempotent `ALTER SYSTEM SET` commands and configuration reload statements.

### 6. Schema Health Auditor (`internal/poller/schema_health.go`)
Inspects relational invariants to prevent locking and performance anomalies:
- **Unindexed Foreign Keys**: Inspects `pg_constraint` against `pg_index` to flag foreign key constraints lacking supporting child indexes. Prevents `ShareRowExclusiveLock` on parent row mutations.
- **Tables Missing Primary Keys**: Identifies user tables without declared primary keys, which hinders logical replication and unique tuple addressing.
- **TOAST Relation Analysis**: Identifies tables where variable-length attributes (JSONB, TEXT) occupy excessive out-of-line storage, advising LZ4 column compression.

### 7. Connection Pool & Saturation Health (`internal/poller/connection_pool.go`)
Monitors client socket lifecycle and saturation:
- Calculates connection saturation against `max_connections`.
- Alerts on long-lived `idle in transaction` sockets holding snapshots that impede autovacuum garbage collection.
- Recommends optimal PgBouncer connection bounds using the $(2 \times \text{vCPUs}) + \text{Disks}$ formula.

### 8. SQL Anti-Pattern Analyzer (`internal/explain/anti_patterns.go`)
Detects query construction flaws that invalidate B-Tree lookups:
- Functions called on indexed columns (e.g. `lower(email) = ...`), synthesizing expression indexes.
- Leading wildcards in `LIKE` / `ILIKE` patterns, recommending PostgreSQL's `pg_trgm` GIN indexes.
- Deep `OFFSET` pagination, suggesting keyset pagination with indexed cursors.

### 9. Web Studio (`web/`)
Built with React 19, TypeScript, Lucide Icons, and Tailwind CSS:
- Zero external SaaS dependencies; operates entirely on the local network or inside secure VPCs.
- Includes a built-in offline telemetry demo mode for development, testing, and evaluation without a live PostgreSQL instance.

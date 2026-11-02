<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="IndexLens Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">IndexLens</h1>

<p align="center">
  <strong>Non-invasive PostgreSQL slow query explainer, autovacuum watchdog, and missing index synthesizer.</strong>
</p>

<p align="center">
  <a href="https://github.com/alexandrmotologa/indexlens/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" alt="License: MIT" /></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version" /></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-19-61DAFB.svg?style=flat-square&logo=react" alt="React 19" /></a>
  <a href="https://www.postgresql.org"><img src="https://img.shields.io/badge/PostgreSQL-14+-4169E1.svg?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL 14+" /></a>
</p>

<p align="center">
  <img src="docs/images/indexlens_demo.gif?raw=true" alt="IndexLens Interactive Studio Demo" width="100%" />
</p>

IndexLens is an open-source, non-invasive PostgreSQL slow query explainer, autovacuum watchdog, and missing index synthesizer. It runs as a self-contained local binary or daemon, reading statistics directly from `pg_stat_statements` and PostgreSQL system catalog views with zero overhead on production workloads.

## Studio Interface & Telemetry Views

Authentic captures from the live IndexLens interactive studio:

### 1. Slow Query Telemetry & Latency Percentiles
Inspect top queries ranked by cumulative runtime, calls, and mean latency. Expand any query row to analyze p50, p90, p95, and p99 percentiles along with latency jitter badges.
<p align="center">
  <img src="docs/images/indexlens_queries.png?raw=true" alt="Slow Query Telemetry" width="100%" />
</p>

### 2. EXPLAIN Execution Flamegraph & Bottleneck Inspector
Paste or load raw JSON execution plans to visualize cost, time distribution, and buffer cache hits across nested plan nodes. Automatic detectors highlight disk spills, sequential scans, and hash batch overflows.
<p align="center">
  <img src="docs/images/indexlens_explain.png?raw=true" alt="EXPLAIN Flamegraph" width="100%" />
</p>

### 3. Missing Index Synthesizer & HypoPG Simulation
Automatically correlate table sequential scans with WHERE and JOIN filters. The synthesizer generates non-blocking `CREATE INDEX CONCURRENTLY` DDL, estimates disk footprint and write overhead, and validates cost drops via HypoPG.
<p align="center">
  <img src="docs/images/indexlens_advisor.png?raw=true" alt="Missing Index Synthesizer" width="100%" />
</p>

### 4. Hardware Calibration & PostgreSQL Tuner
Calibrate critical memory parameters (`shared_buffers`, `work_mem`, `effective_cache_size`, `random_page_cost`, worker counts) based on available RAM, vCPU cores, and storage media. Generates copyable `ALTER SYSTEM` statements.
<p align="center">
  <img src="docs/images/indexlens_tuner.png?raw=true" alt="Hardware Configuration Tuner" width="100%" />
</p>

### 5. Schema Health & Relational Anti-Pattern Auditor
Detect unindexed foreign keys that cause full table locks on cascading updates or deletes, find tables missing primary keys, and identify high out-of-line TOAST storage.
<p align="center">
  <img src="docs/images/indexlens_schema.png?raw=true" alt="Schema Health Auditor" width="100%" />
</p>

### 6. Connection Pool & Saturation Health
Track active and idle connection counts, receive early warnings on leaked `idle in transaction` sessions, and calculate recommended PgBouncer pool sizing using physical CPU core counts.
<p align="center">
  <img src="docs/images/indexlens_pool.png?raw=true" alt="Connection Pool Health" width="100%" />
</p>

### 7. Table Bloat & Autovacuum Guard
Monitor dead tuple ratios, estimate wasted disk space from MVCC update churn, track running autovacuum workers, and alert before Transaction ID (TxID) Wraparound limits are reached.
<p align="center">
  <img src="docs/images/indexlens_bloat.png?raw=true" alt="Table Bloat and Autovacuum Guard" width="100%" />
</p>

### 8. Responsive Mobile Experience
Full responsive navigation with horizontal scroll tab bar and compact telemetry cards for on-call triage on mobile and tablet devices.
<p align="center">
  <img src="docs/images/indexlens_mobile.png?raw=true" alt="IndexLens Mobile View" width="420" />
</p>

## Key Capabilities

- **Non-Invasive Telemetry Polling**: Connects using standard read-only credentials. Inspects `pg_stat_statements`, `pg_stat_user_tables`, `pg_statio_user_tables`, and `pg_statio_user_indexes`.
- **Visual EXPLAIN Flamegraph**: Parses `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` execution plans into interactive hierarchical flamegraphs. Highlights buffer cache hits, read latency, and node execution time.
- **Bottleneck Detection**: Automatically detects heavy sequential table scans, disk spills (`Sort Method: external merge Disk`), hash join batch overflow, and stale planner statistics.
- **Missing Index Synthesizer**: Correlates sequential scans with active query filter patterns. Enforces B-Tree ordering rules (equality columns first, range columns next, sort keys last) and synthesizes non-blocking `CREATE INDEX CONCURRENTLY` DDL statements. Includes index disk footprint and write-penalty calculations.
- **Unused & Redundant Index Cleanup**: Identifies indexes with zero or negligible scans and redundant prefix indexes, generating non-blocking `DROP INDEX CONCURRENTLY IF EXISTS` statements.
- **Plan Diff & Regression Analyzer**: Compares baseline vs current execution plans side-by-side, isolating duration shifts, buffer read spikes, and node type changes.
- **Live Lock Contention Tracker**: Inspects `pg_stat_activity` and `pg_locks` to map root blocker transactions, blocked waiting queries, and idle-in-transaction connections.
- **Table Bloat & Autovacuum Watchdog**: Calculates dead tuple ratios, estimates wasted disk space from MVCC update churn, tracks running `pg_stat_progress_vacuum` phases, and alerts before Transaction ID (TxID) Wraparound limits are reached.
- **HypoPG In-Memory Verification**: Simulates hypothetical indexes using the `hypopg` extension to prove planner cost reduction before running physical index builds.
- **Hardware & PostgreSQL Configuration Tuner**: Calibrates `shared_buffers`, `work_mem`, `effective_cache_size`, `random_page_cost`, and worker counts against physical RAM, vCPU cores, and disk speed, outputting an `ALTER SYSTEM SET` script.
- **Schema Health & Relational Anti-Pattern Auditor**: Detects unindexed foreign keys that cause table locks during deletes, identifies tables missing primary keys, and flags high out-of-line TOAST relations.
- **Query Latency Distribution & Jitter Detection**: Calculates p50, p90, p95, and p99 latency percentiles alongside the coefficient of variation to spot volatile queries suffering from lock or disk jitter.
- **Connection Pool & Saturation Health**: Tracks active vs idle connection ratios, flags long-lived `idle in transaction` queries, and calculates optimal PgBouncer pool sizing using physical core counts.
- **SQL Anti-Pattern Rewriter**: Identifies function calls on indexed columns (`lower(email)`), unindexed leading wildcards (`LIKE '%term%'`), and deep `OFFSET` pagination, providing concrete rewrite patterns and expression index DDL.
- **Local-First Web Studio & CI/CD Gate**: Interactive React 19 interface served directly from the Go binary without external cloud dependencies. Includes an automated `check` command for CI/CD pipelines and 1-click Markdown report exports.

## Architecture

```
                      +------------------------------------+
                      |    PostgreSQL Database Instance    |
                      |   pg_stat_statements / pg_statio   |
                      +-----------------+------------------+
                                        | Read-Only Polling
                                        v
+----------------------------------------------------------------------------+
|                             INDEXLENS DAEMON                               |
|                                                                            |
|  +-------------------+  +---------------------+  +----------------------+  |
|  |  Poller Client    |  |  EXPLAIN AST Parser |  |  Index Synthesizer   |  |
|  |  (pgx connection) |  |  (Flamegraph Model) |  |  (B-Tree Ranker)     |  |
|  +---------+---------+  +----------+----------+  +----------+-----------+  |
|            |                       |                        |              |
|            +-----------------------+------------------------+              |
|                                    |                                       |
|                         +----------v-----------+                           |
|                         |  Bloat & TxID Guard  |                           |
|                         +----------+-----------+                           |
|                                    |                                       |
|                         +----------v-----------+                           |
|                         |  Embedded Web Studio |                           |
|                         |  (Go + React 19)     |                           |
|                         +----------+-----------+                           |
+------------------------------------|---------------------------------------+
                                     | HTTP :9920
                                     v
                       +---------------------------+
                       |    Browser Dashboard      |
                       +---------------------------+
```

## Quick Start

### 1. Installation

Install via Go:

```bash
go install github.com/alexandrmotologa/indexlens/cmd/indexlens@latest
```

Or build from source:

```bash
git clone https://github.com/alexandrmotologa/indexlens.git
cd indexlens
go build -o indexlens ./cmd/indexlens
```

### 2. Running in Demo Mode (No Database Required)

To evaluate IndexLens without an active PostgreSQL connection:

```bash
./indexlens serve --demo --port 9920
```

Open your browser at `http://localhost:9920` to inspect simulated production workloads, execution flamegraphs, and index recommendations.

### 3. Connecting to PostgreSQL

Ensure `pg_stat_statements` is enabled in your database. See [docs/POSTGRESQL_SETUP.md](docs/POSTGRESQL_SETUP.md) for step-by-step instructions and permissions.

```bash
./indexlens serve --dsn "postgres://monitor_user:secret@localhost:5432/production_db?sslmode=prefer" --port 9920
```

### 4. CLI Subcommands

IndexLens provides a suite of terminal commands for diagnostics and automation:

```bash
# Start the web studio and REST API
./indexlens serve --dsn "postgres://..." --port 9920

# Analyze a JSON plan file or run a live query EXPLAIN
./indexlens explain --file plan.json
./indexlens explain --dsn "postgres://..." "SELECT * FROM orders WHERE tenant_id = 42"

# Review missing index advice
./indexlens advice --dsn "postgres://..."

# Detect unused and redundant indexes
./indexlens unused --dsn "postgres://..."

# Inspect real-time lock contention and blocking query trees
./indexlens locks --dsn "postgres://..."

# Calculate hardware-tailored postgresql.conf settings and ALTER SYSTEM script
./indexlens tune --ram-gb 32 --vcpus 8 --disk NVMe --workload OLTP

# Audit unindexed foreign keys, missing primary keys, and TOAST bloat
./indexlens schema --dsn "postgres://..."

# Inspect connection saturation, idle-in-tx leaks, and PgBouncer sizing
./indexlens pool --dsn "postgres://..." --vcpus 8

# Estimate table dead tuples and disk bloat
./indexlens bloat --dsn "postgres://..."

# Run CI/CD quality gate check with failure thresholds
./indexlens check --dsn "postgres://..." --fail-on-bloat 35 --fail-on-cache 90 --fail-on-wraparound 75

# Export a complete Markdown audit report
./indexlens report --dsn "postgres://..." --output audit_report.md
```

## REST API Reference

When running `indexlens serve`, the internal API exposes the following endpoints:

| Endpoint | Method | Description |
|---|---|---|
| `/api/health` | GET | Health status and connection state |
| `/api/overview` | GET | Global metrics: cache hit ratio, active locks, bloat stats |
| `/api/queries` | GET | Top slow queries from `pg_stat_statements` with sorting and filtering |
| `/api/tables` | GET | Table statistics, sequential vs index scan counts, and dead tuple ratios |
| `/api/cache` | GET | Global heap/index buffer cache hit ratios and sub-optimal relations |
| `/api/explain` | POST | Ingests JSON execution plans and returns flamegraphs and bottlenecks |
| `/api/explain/diff` | POST | Compares baseline vs current execution plans and computes latency deltas |
| `/api/indexes/advisor` | GET | Synthesized missing index suggestions with non-blocking DDL |
| `/api/indexes/unused` | GET | Zero-scan and redundant indexes with safe DROP statements |
| `/api/indexes/simulate` | POST | Tests index planner selection in memory via HypoPG |
| `/api/activity/locks` | GET | Active connections, wait events, and blocking dependency trees |
| `/api/autovacuum` | GET | Running autovacuum workers and TxID wraparound safety horizon |
| `/api/bloat` | GET | Table and index bloat estimates |
| `/api/config/tuning` | GET/POST | Recommended postgresql.conf settings and copyable ALTER SYSTEM script |
| `/api/schema/health` | GET | Audit of unindexed foreign keys, tables missing primary keys, and TOAST bloat |
| `/api/connections/pool` | GET | Saturation analysis, idle-in-tx leak detection, and PgBouncer sizing |
| `/api/queries/anti-patterns` | POST | Identifies SQL anti-patterns (e.g. lower(col), leading wildcard, deep offset) |
| `/api/report/markdown` | GET | Generates downloadable Markdown incident / audit report |

## Documentation

- [PostgreSQL Setup & Permissions](docs/POSTGRESQL_SETUP.md)
- [Architecture & Data Flows](docs/ARCHITECTURE.md)
- [Missing Index Synthesis Heuristics](docs/INDEX_SYNTHESIS_RULES.md)
- [Autovacuum & Table Bloat Mechanics](docs/AUTOVACUUM_AND_BLOAT.md)

## License

MIT License. See [LICENSE](LICENSE) for details.

# IndexLens

IndexLens is a non-invasive PostgreSQL slow query explainer, autovacuum watchdog, and missing index synthesizer. It runs as a self-contained local binary or daemon, reading statistics directly from `pg_stat_statements` and PostgreSQL system catalog views with zero overhead on production workloads.

## Key Capabilities

- **Non-Invasive Telemetry Polling**: Connects using standard read-only credentials. Inspects `pg_stat_statements`, `pg_stat_user_tables`, `pg_statio_user_tables`, and `pg_statio_user_indexes`.
- **Visual EXPLAIN Flamegraph**: Parses `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` execution plans into interactive hierarchical flamegraphs. Highlights buffer cache hits, read latency, and node execution time.
- **Bottleneck Detection**: Automatically detects heavy sequential table scans, disk spills (`Sort Method: external merge Disk`), hash join batch overflow, and stale planner statistics.
- **Missing Index Synthesizer**: Correlates sequential scans with active query filter patterns. Enforces B-Tree ordering rules (equality columns first, range columns next, sort keys last) and synthesizes non-blocking `CREATE INDEX CONCURRENTLY` DDL statements.
- **Table Bloat & Autovacuum Watchdog**: Calculates dead tuple ratios, estimates wasted disk space from MVCC update churn, tracks running `pg_stat_progress_vacuum` phases, and alerts before Transaction ID (TxID) Wraparound limits are reached.
- **Local-First Web Studio**: Interactive React 19 interface served directly from the Go binary without external cloud dependencies or telemetry tracking. Includes an offline demo mode for testing without a live database.

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

Analyze an EXPLAIN plan directly in your terminal:

```bash
# Analyze a JSON plan file
./indexlens explain --file plan.json

# Check table bloat directly from the CLI
./indexlens bloat --dsn "postgres://monitor_user:secret@localhost:5432/production_db"

# Review missing index advice from the terminal
./indexlens advice --dsn "postgres://monitor_user:secret@localhost:5432/production_db"
```

## REST API Reference

When running `indexlens serve`, the internal API exposes the following endpoints:

| Endpoint | Method | Description |
|---|---|---|
| `/api/health` | GET | Health status and connection state |
| `/api/overview` | GET | Global metrics: buffer cache hit ratio, active connections, bloat stats |
| `/api/queries` | GET | Top slow queries from `pg_stat_statements` with sorting and filtering |
| `/api/tables` | GET | Table statistics, sequential vs index scan counts, and dead tuple ratios |
| `/api/explain` | POST | Ingests raw JSON execution plans and returns flamegraph nodes and bottlenecks |
| `/api/indexes/advisor` | GET | Synthesized missing index suggestions with non-blocking DDL |
| `/api/autovacuum` | GET | Running autovacuum workers and TxID wraparound safety horizon |

## Documentation

- [PostgreSQL Setup & Permissions](docs/POSTGRESQL_SETUP.md)
- [Architecture & Data Flows](docs/ARCHITECTURE.md)
- [Missing Index Synthesis Heuristics](docs/INDEX_SYNTHESIS_RULES.md)
- [Autovacuum & Table Bloat Mechanics](docs/AUTOVACUUM_AND_BLOAT.md)

## License

MIT License. See [LICENSE](LICENSE) for details.

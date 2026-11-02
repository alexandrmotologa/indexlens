# PostgreSQL Setup and Permissions Guide

IndexLens operates non-invasively by reading PostgreSQL internal catalog views and statistics counters. It does not alter your application schema or execute mutating queries.

## 1. Enabling pg_stat_statements

IndexLens relies on the `pg_stat_statements` extension to collect query performance metrics such as execution count, total duration, mean latency, and disk buffer usage.

### Edit postgresql.conf

Ensure `pg_stat_statements` is added to `shared_preload_libraries`:

```ini
shared_preload_libraries = 'pg_stat_statements'

# Recommended tracking settings
pg_stat_statements.track = top
pg_stat_statements.max = 10000
pg_stat_statements.track_utility = off
```

Restart your PostgreSQL server after modifying `shared_preload_libraries`.

### Create Extension

Connect to your database as a superuser and run:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

## 2. Creating a Dedicated Read-Only User

Create an unprivileged user specifically for IndexLens telemetry collection:

```sql
-- 1. Create monitoring user
CREATE USER indexlens_monitor WITH PASSWORD 'your_secure_password';

-- 2. Grant connection rights to the database
GRANT CONNECT ON DATABASE your_database TO indexlens_monitor;

-- 3. Grant schema usage
GRANT USAGE ON SCHEMA public TO indexlens_monitor;

-- 4. Grant read permissions on system statistics (PostgreSQL 10+)
GRANT pg_read_all_stats TO indexlens_monitor;

-- 5. If running on PostgreSQL 14+, allow viewing active statements
GRANT pg_read_all_data TO indexlens_monitor;
```

## 3. Connection String Format

Provide the read-only credentials to IndexLens via the `--dsn` flag or `DATABASE_URL` environment variable:

```bash
# Standard DSN
postgres://indexlens_monitor:your_secure_password@localhost:5432/your_database?sslmode=prefer

# Starting IndexLens server
indexlens serve --dsn "postgres://indexlens_monitor:your_secure_password@localhost:5432/your_database?sslmode=prefer" --port 9920
```

## 4. Required Catalog Views

IndexLens inspects the following views:

| View | Purpose |
|---|---|
| `pg_stat_statements` | Aggregated query latencies, call counts, and buffer hits |
| `pg_stat_user_tables` | Table sequential scans, index scans, and live/dead tuple counts |
| `pg_statio_user_tables` | Heap block reads vs hits for table cache ratio |
| `pg_statio_user_indexes` | Index block reads vs hits for index cache ratio |
| `pg_stat_progress_vacuum` | Real-time progress and phases of running vacuum workers |
| `pg_database` | Database age and transaction freeze status for wraparound protection |
| `pg_class` & `pg_stats` | Tuple header sizes and statistical estimates for bloat calculation |

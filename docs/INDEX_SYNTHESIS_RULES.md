# Missing Index Synthesis Heuristics

IndexLens identifies tables suffering from excessive sequential scans and inspects query filter predicates to synthesize non-blocking `CREATE INDEX CONCURRENTLY` statements.

## 1. Candidate Identification

IndexLens flags a table for missing index analysis when:
1. `seq_scan > 50` and `seq_tup_read > 10,000`.
2. The ratio of sequential scans to total scans (`seq_scan / (seq_scan + idx_scan)`) exceeds 0.25 on tables with more than 5,000 live rows.
3. Slow queries recorded in `pg_stat_statements` filter on the table without using an index scan in their execution plan.

## 2. B-Tree Multi-Column Ordering Rules

When constructing composite indexes, column order determines index selectivity and usability. IndexLens applies the standard B-Tree indexing rules:

```
[ Equality Predicates ]  -->  [ Inequality / Range Predicates ]  -->  [ ORDER BY / Sort Columns ]
      col1 = val                    col2 >= val                          ORDER BY col3 ASC
     col1 IS NULL                   col2 BETWEEN a AND b
```

### Rule 1: Equality Columns First
Columns evaluated with `=` or `IS NULL` reduce the search space to a discrete leaf point. These columns are always placed leftmost in the index definition:
```sql
-- Query: WHERE tenant_id = 42 AND created_at >= '2026-01-01'
-- Optimal index column order:
(tenant_id, created_at)
```

### Rule 2: Range and Inequality Predicates Next
Columns filtered using `<`, `>`, `<=`, `>=`, `BETWEEN`, `IN`, or `LIKE 'prefix%'` allow binary search down to the boundary, followed by sequential leaf scanning. Only one range column can effectively bound the index traversal. Place range columns immediately after equality columns.

### Rule 3: Sort Keys Last (Index-Organized Sorting)
If the query contains `ORDER BY col DESC LIMIT N`, appending `col DESC` to the index eliminates an explicit disk/memory sort node in the execution plan:
```sql
-- Query: WHERE status = 'active' ORDER BY updated_at DESC LIMIT 50
CREATE INDEX CONCURRENTLY idx_items_status_updated 
ON items (status, updated_at DESC);
```

## 3. Partial Index Detection

Creating indexes on full tables wastes disk space when queries target a specific sub-population:
- If a query always filters by `deleted_at IS NULL`, IndexLens recommends a partial index:
  ```sql
  CREATE INDEX CONCURRENTLY idx_users_email_active 
  ON users (email) 
  WHERE deleted_at IS NULL;
  ```
- If a table has an enum or status column where 95% of rows are `completed` and queries only search for `pending` or `failed`:
  ```sql
  CREATE INDEX CONCURRENTLY idx_jobs_pending 
  ON background_jobs (id) 
  WHERE status = 'pending';
  ```

## 4. Production Safety Rules

1. **Non-Blocking DDL**: Every suggested statement includes `CONCURRENTLY` to avoid taking an exclusive table write lock (`ShareLock`), ensuring ongoing transactions are not blocked:
   ```sql
   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_... ON ... (...);
   ```
2. **Transaction Blocks**: `CREATE INDEX CONCURRENTLY` cannot run inside an explicit transaction block (`BEGIN ... COMMIT`).
3. **Invalid Indexes Check**: If a concurrent index build is cancelled or interrupted by an error, PostgreSQL leaves an invalid index marked `INVALID`. IndexLens checks for invalid indexes in `pg_index` and alerts administrators.

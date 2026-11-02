# Autovacuum Mechanics, Table Bloat, and Wraparound Guard

IndexLens continuously monitors table dead tuples, estimated disk bloat, autovacuum worker activity, and transaction ID consumption.

## 1. PostgreSQL MVCC and Bloat

PostgreSQL implements Multi-Version Concurrency Control (MVCC). When a row is updated or deleted:
1. `UPDATE`: PostgreSQL marks the old row version as expired (`xmax` is set) and inserts a new physical tuple.
2. `DELETE`: The tuple is marked dead.
3. The disk space occupied by the dead tuple cannot be reused for new rows until `VACUUM` processes the table page.

When write rates outpace autovacuum workers, dead tuples accumulate. This creates table and index bloat:
- Table files (`base/<db_oid>/<rel_oid>`) expand unnecessarily on disk.
- Sequential scans must read empty or obsolete pages into memory, evicting active data from the PostgreSQL `shared_buffers` cache.
- Buffer cache hit ratios drop from > 99% down to 80% or lower.

## 2. Table Bloat Estimation Formula

IndexLens calculates dead tuple ratios directly from `pg_stat_user_tables`:

$$\text{DeadTupleRatio} = \frac{\text{n\_dead\_tup}}{\text{n\_live\_tup} + \text{n\_dead\_tup}} \times 100\%$$

For disk byte bloat estimation, IndexLens inspects `pg_class`, `pg_stats`, and column average widths to estimate physical page density:

1. **Header Size**: 24 bytes per heap tuple header + 4-byte line pointer (`ItemIdData`).
2. **Page Fill**: Default 8KB (8192 bytes) page with page header overhead (24 bytes).
3. **Expected vs Actual Pages**:
   $$\text{ExpectedPages} = \lceil \frac{\text{reltuples} \times \text{AvgTupleWidth}}{\text{PageUsableBytes}} \rceil$$
   $$\text{BloatBytes} = \max(0, (\text{relpages} - \text{ExpectedPages}) \times 8192)$$

## 3. Autovacuum Worker Monitoring

PostgreSQL limits concurrent autovacuum workers via `autovacuum_max_workers` (default 3). IndexLens queries `pg_stat_progress_vacuum` to expose:
- `pid`: Worker process ID.
- `phase`: Current operation (`initializing`, `scanning heap`, `vacuuming indexes`, `vacuuming heap`, `cleaning up index`, `truncating heap`, `performing final cleanup`).
- `heap_blks_scanned` / `heap_blks_total`: Percentage of heap scanned.
- `index_vacuum_count`: Number of index vacuum passes completed.

## 4. Transaction ID (TxID) Wraparound Prevention

PostgreSQL uses 32-bit transaction IDs, allowing roughly 4.29 billion transactions before numbers wrap around. To avoid confusing old past transactions with future transactions, PostgreSQL treats numbers within a 2-billion window as past transactions.

If a database approaches 2 billion transactions without vacuum freezing old tuples, PostgreSQL enters read-only emergency mode to prevent data corruption.

IndexLens queries the age of each database's frozen transaction ID:

```sql
SELECT datname, age(datfrozenxid) AS txid_age 
FROM pg_database 
WHERE datallowconn = true;
```

It compares `txid_age` against `autovacuum_freeze_max_age` (default 200,000,000 transactions):
- **Normal**: `txid_age < 100,000,000` (Safety horizon > 50%).
- **Warning**: `100,000,000 <= txid_age < 150,000,000`.
- **Critical Alert**: `txid_age >= 150,000,000` (Autovacuum is aggressively freezing or failing to keep pace).

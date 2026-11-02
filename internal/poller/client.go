package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps a PostgreSQL connection pool for read-only telemetry queries.
type Client struct {
	pool    *pgxpool.Pool
	dsn     string
	dbName  string
	version string
}

// NewClient initializes a connection pool with conservative read-only settings.
func NewClient(ctx context.Context, dsn string) (*Client, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL DSN: %w", err)
	}

	// Keep pool small and non-invasive: max 4 connections, quick idle cleanup
	config.MaxConns = 4
	config.MinConns = 1
	config.MaxConnIdleTime = 2 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 15 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	c := &Client{
		pool: pool,
		dsn:  dsn,
	}

	// Read version and current database
	_ = pool.QueryRow(ctx, "SELECT version(), current_database()").Scan(&c.version, &c.dbName)

	return c, nil
}

// Close terminates connection pool resources.
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Ping checks if the database is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("connection pool is closed")
	}
	return c.pool.Ping(ctx)
}

// HasExtension checks if a specified PostgreSQL extension is installed and available.
func (c *Client) HasExtension(ctx context.Context, extName string) (bool, error) {
	query := "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = $1)"
	var exists bool
	err := c.pool.QueryRow(ctx, query, extName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check extension %s: %w", extName, err)
	}
	return exists, nil
}

// Pool returns the underlying pgxpool instance.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// GetOverview collects active connections and database health indicators.
func (c *Client) GetOverview(ctx context.Context) (model.SystemOverview, error) {
	overview := model.SystemOverview{
		Connected:       true,
		DemoMode:        false,
		DatabaseVersion: c.version,
		DatabaseName:    c.dbName,
	}

	// Query active connections and max connections
	var activeConns, maxConns int
	err := c.pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state != 'idle'").Scan(&activeConns)
	if err == nil {
		overview.ActiveConnections = activeConns
	}

	var maxConnsStr string
	err = c.pool.QueryRow(ctx, "SHOW max_connections").Scan(&maxConnsStr)
	if err == nil {
		_, _ = fmt.Sscanf(maxConnsStr, "%d", &maxConns)
		overview.MaxConnections = maxConns
	}

	// Active autovacuum workers count
	var activeAutovacuums int
	_ = c.pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_progress_vacuum").Scan(&activeAutovacuums)
	overview.ActiveAutovacuums = activeAutovacuums

	// Total queries tracked
	hasExt, _ := c.HasExtension(ctx, "pg_stat_statements")
	if hasExt {
		var totalQueries int
		_ = c.pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_statements").Scan(&totalQueries)
		overview.TotalQueriesTracked = totalQueries
	}

	return overview, nil
}

package poller

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// CalculateTuning computes recommended PostgreSQL settings based on hardware profile.
func CalculateTuning(profile model.TuningProfile, currentSettings map[string]string) model.TuningReport {
	ramGB := profile.TotalRAMGB
	if ramGB <= 0 {
		ramGB = 8.0 // default fallback
	}
	vcpus := profile.VCPUs
	if vcpus <= 0 {
		vcpus = 4
	}
	disk := strings.ToUpper(profile.DiskType)
	if disk == "" {
		disk = "SSD"
	}
	workload := strings.ToUpper(profile.WorkloadType)
	if workload == "" {
		workload = "OLTP"
	}

	report := model.TuningReport{
		Profile:  profile,
		Settings: make([]model.ConfigSetting, 0),
	}

	// 1. shared_buffers: 25% of RAM, capped around 16GB
	sharedBuffersMB := int64(ramGB * 1024 * 0.25)
	if sharedBuffersMB > 16384 {
		sharedBuffersMB = 16384
	}
	if sharedBuffersMB < 128 {
		sharedBuffersMB = 128
	}

	// 2. effective_cache_size: 75% of RAM (or 50% for mixed)
	cacheRatio := 0.75
	if workload == "MIXED" {
		cacheRatio = 0.50
	}
	effectiveCacheMB := int64(ramGB * 1024 * cacheRatio)

	// 3. maintenance_work_mem: 10% of RAM, max 2048MB
	maintWorkMemMB := int64(ramGB * 1024 * 0.10)
	if maintWorkMemMB > 2048 {
		maintWorkMemMB = 2048
	}
	if maintWorkMemMB < 64 {
		maintWorkMemMB = 64
	}

	// 4. work_mem: based on connections (assume 100 max_connections typical)
	// (Total RAM - shared_buffers) / (max_connections * 3)
	availableForWorkMB := (ramGB * 1024) - float64(sharedBuffersMB)
	workMemMB := int64(availableForWorkMB / 300.0)
	if workload == "DW" {
		workMemMB = int64(availableForWorkMB / 100.0)
	}
	if workMemMB < 4 {
		workMemMB = 4
	}
	if workMemMB > 256 {
		workMemMB = 256
	}

	// 5. wal_buffers: 16MB or 3% of shared_buffers
	walBuffersMB := int64(float64(sharedBuffersMB) * 0.03)
	if walBuffersMB > 64 {
		walBuffersMB = 64
	}
	if walBuffersMB < 16 {
		walBuffersMB = 16
	}

	// 6. random_page_cost & effective_io_concurrency
	randomPageCost := "1.1"
	effectiveIOConcurrency := "200"
	if disk == "HDD" {
		randomPageCost = "4.0"
		effectiveIOConcurrency = "2"
	}

	// 7. Parallel workers
	maxWorkers := vcpus
	maxParallelGather := int(math.Max(1, float64(vcpus)/2))
	if maxParallelGather > 4 && workload != "DW" {
		maxParallelGather = 4
	}
	maxParallelMaint := int(math.Max(1, float64(vcpus)/2))
	if maxParallelMaint > 4 {
		maxParallelMaint = 4
	}

	// Definitions of targets
	type targetDef struct {
		Name     string
		RecValue string
		Unit     string
		Category string
		Restart  bool
		Reason   string
	}

	targets := []targetDef{
		{
			Name:     "shared_buffers",
			RecValue: fmt.Sprintf("%dMB", sharedBuffersMB),
			Unit:     "MB",
			Category: "Memory",
			Restart:  true,
			Reason:   "Allocates 25% of system RAM to PostgreSQL dedicated page cache, preventing kernel context-switch stalls.",
		},
		{
			Name:     "effective_cache_size",
			RecValue: fmt.Sprintf("%dMB", effectiveCacheMB),
			Unit:     "MB",
			Category: "Memory",
			Restart:  false,
			Reason:   "Planner assumption for total memory available for disk caching (PostgreSQL buffer cache + OS page cache).",
		},
		{
			Name:     "work_mem",
			RecValue: fmt.Sprintf("%dMB", workMemMB),
			Unit:     "MB",
			Category: "Memory",
			Restart:  false,
			Reason:   "Allocated per sort/hash operation before spilling temporary files to disk. Sized for concurrency without OOM risk.",
		},
		{
			Name:     "maintenance_work_mem",
			RecValue: fmt.Sprintf("%dMB", maintWorkMemMB),
			Unit:     "MB",
			Category: "Memory",
			Restart:  false,
			Reason:   "Accelerates VACUUM, CREATE INDEX, and ALTER TABLE foreign key validations.",
		},
		{
			Name:     "checkpoint_completion_target",
			RecValue: "0.9",
			Unit:     "ratio",
			Category: "WAL",
			Restart:  false,
			Reason:   "Spreads dirty page write I/O across 90% of the checkpoint duration, eliminating I/O spikes.",
		},
		{
			Name:     "wal_buffers",
			RecValue: fmt.Sprintf("%dMB", walBuffersMB),
			Unit:     "MB",
			Category: "WAL",
			Restart:  true,
			Reason:   "Dedicates buffer space for unwritten WAL data, avoiding disk write bottlenecks during concurrent transactions.",
		},
		{
			Name:     "random_page_cost",
			RecValue: randomPageCost,
			Unit:     "ratio",
			Category: "Query Planner",
			Restart:  false,
			Reason:   "Cost estimate for non-sequential disk reads. 1.1 reflects fast flash/NVMe storage random access.",
		},
		{
			Name:     "effective_io_concurrency",
			RecValue: effectiveIOConcurrency,
			Unit:     "requests",
			Category: "Query Planner",
			Restart:  false,
			Reason:   "Allows asynchronous prefetching of heap pages during bitmap index scans.",
		},
		{
			Name:     "max_worker_processes",
			RecValue: fmt.Sprintf("%d", maxWorkers),
			Unit:     "workers",
			Category: "Parallelism",
			Restart:  true,
			Reason:   "Caps background workers to match system physical cores for optimal parallel query execution.",
		},
		{
			Name:     "max_parallel_workers_per_gather",
			RecValue: fmt.Sprintf("%d", maxParallelGather),
			Unit:     "workers",
			Category: "Parallelism",
			Restart:  false,
			Reason:   "Maximum parallel threads spawned for a single parallel query node (e.g. Parallel Seq Scan).",
		},
		{
			Name:     "max_parallel_maintenance_workers",
			RecValue: fmt.Sprintf("%d", maxParallelMaint),
			Unit:     "workers",
			Category: "Parallelism",
			Restart:  false,
			Reason:   "Parallel threads for CREATE INDEX and autovacuum operations.",
		},
	}

	var alterScript strings.Builder
	alterScript.WriteString("-- IndexLens Generated PostgreSQL Hardware Tuning Script\n")
	alterScript.WriteString(fmt.Sprintf("-- Profile: %.0f GB RAM, %d vCPUs, %s storage, %s workload\n\n", ramGB, vcpus, disk, workload))

	for _, t := range targets {
		curr := currentSettings[t.Name]
		if curr == "" {
			curr = "default"
		}

		status := "OPTIMAL"
		if strings.ToLower(curr) != strings.ToLower(t.RecValue) && curr != "default" {
			status = "ATTENTION"
			report.AttentionCount++
		} else if curr == "default" {
			status = "ATTENTION"
			report.AttentionCount++
		} else {
			report.OptimalCount++
		}

		ddl := fmt.Sprintf("ALTER SYSTEM SET %s = '%s';", t.Name, t.RecValue)
		alterScript.WriteString(fmt.Sprintf("%s -- %s\n", ddl, t.Reason))

		report.Settings = append(report.Settings, model.ConfigSetting{
			Name:             t.Name,
			CurrentValue:     curr,
			RecommendedValue: t.RecValue,
			Unit:             t.Unit,
			Category:         t.Category,
			RestartRequired:  t.Restart,
			Reason:           t.Reason,
			Status:           status,
			AlterDDL:         ddl,
		})
	}

	alterScript.WriteString("\n-- To apply changes without restart (for parameters that do not require restart):\n")
	alterScript.WriteString("SELECT pg_reload_conf();\n")

	report.AlterScript = alterScript.String()
	return report
}

// GetLiveSettings queries current values from pg_settings for tuned parameters.
func (c *Client) GetLiveSettings(ctx context.Context) (map[string]string, error) {
	query := `
		SELECT name, setting || COALESCE(unit, '') AS val
		FROM pg_settings
		WHERE name IN (
			'shared_buffers', 'effective_cache_size', 'work_mem', 'maintenance_work_mem',
			'checkpoint_completion_target', 'wal_buffers', 'random_page_cost',
			'effective_io_concurrency', 'max_worker_processes',
			'max_parallel_workers_per_gather', 'max_parallel_maintenance_workers'
		)
	`
	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to read pg_settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err == nil {
			settings[name] = val
		}
	}
	return settings, rows.Err()
}

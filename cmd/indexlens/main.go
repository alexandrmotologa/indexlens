package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexandrmotologa/indexlens/internal/explain"
	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
	"github.com/alexandrmotologa/indexlens/internal/server"
	"github.com/alexandrmotologa/indexlens/internal/synthesizer"
	"github.com/alexandrmotologa/indexlens/internal/watchdog"
)

const version = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "serve":
		runServe(args)
	case "explain":
		runExplain(args)
	case "bloat":
		runBloat(args)
	case "advice":
		runAdvice(args)
	case "unused":
		runUnused(args)
	case "locks":
		runLocks(args)
	case "check":
		runCheck(args)
	case "report":
		runReport(args)
	case "version", "-v", "--version":
		fmt.Printf("IndexLens v%s (PostgreSQL Diagnostics & Index Synthesizer)\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	help := `IndexLens: Non-Invasive PostgreSQL Slow Query Explainer & Missing Index Synthesizer

Usage:
  indexlens <command> [flags] [arguments]

Core Commands:
  serve     Start the interactive web studio and REST API
  explain   Parse and diagnose an EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) plan
  bloat     Estimate table dead tuples and disk bloat
  advice    Synthesize non-blocking CREATE INDEX CONCURRENTLY recommendations
  unused    Detect zero-scan and redundant indexes with safe DROP DDL
  locks     Inspect real-time lock contention and blocking processes
  check     Automated CI/CD health verification with exit codes
  report    Generate a comprehensive Markdown audit report
  version   Display version information
  help      Show help overview

Run 'indexlens <command> --help' for details on a specific command.
`
	fmt.Print(help)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL read-only connection string")
	port := fs.Int("port", 9920, "HTTP server listening port")
	demo := fs.Bool("demo", false, "Run in offline demo mode with simulated production telemetry")
	_ = fs.Parse(args)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var client *poller.Client
	if !*demo && *dsn != "" {
		fmt.Printf("[IndexLens] Connecting to PostgreSQL at %s...\n", maskDSN(*dsn))
		var err error
		client, err = poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Printf("[IndexLens] Warning: connection failed (%v). Falling back to demo mode.\n", err)
			*demo = true
		} else {
			defer client.Close()
			hasExt, _ := client.HasExtension(ctx, "pg_stat_statements")
			if !hasExt {
				fmt.Println("[IndexLens] Notice: pg_stat_statements extension not found. Some query telemetry will be limited.")
			}
			fmt.Println("[IndexLens] Successfully connected with read-only privileges.")
		}
	} else {
		*demo = true
		fmt.Println("[IndexLens] Starting in autonomous demo mode with simulated production telemetry.")
	}

	srv := server.NewServer(client, *demo)
	srv.RegisterStaticRoutes()

	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("[IndexLens] Web Studio live at: http://localhost:%d\n", *port)
		fmt.Println("[IndexLens] Press Ctrl+C to terminate.")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "[IndexLens] Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	fmt.Println("\n[IndexLens] Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	fmt.Println("[IndexLens] Stopped.")
}

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	filePath := fs.String("file", "", "Path to JSON execution plan file")
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN to run live query EXPLAIN")
	_ = fs.Parse(args)

	var planJSON []byte

	if *filePath != "" {
		data, err := os.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading plan file: %v\n", err)
			os.Exit(1)
		}
		planJSON = data
	} else if len(fs.Args()) > 0 && *dsn != "" {
		query := fs.Args()[0]
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		var planStr string
		err = client.Pool().QueryRow(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "+query).Scan(&planStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run EXPLAIN: %v\n", err)
			os.Exit(1)
		}
		planJSON = []byte(planStr)
	} else {
		fmt.Println("[IndexLens] Analyzing demo plan (orders table query):")
		res := model.GetDemoExplainResult()
		printExplainResult(&res)
		return
	}

	res, err := explain.ParsePlanJSON(planJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing plan: %v\n", err)
		os.Exit(1)
	}

	printExplainResult(res)
}

func printExplainResult(res *model.ExplainAnalysisResult) {
	fmt.Printf("\n=== Execution Plan Summary ===\n")
	fmt.Printf("Total Duration : %.2f ms\n", res.TotalExecutionTime)
	fmt.Printf("Planning Time  : %.2f ms\n", res.PlanningTime)
	fmt.Printf("Buffer Cache   : %.1f%% (%d hit, %d read from disk)\n",
		res.CacheHitRatio, res.TotalBuffersHit, res.TotalBuffersRead)

	if len(res.Bottlenecks) > 0 {
		fmt.Printf("\n=== Identified Bottlenecks (%d) ===\n", len(res.Bottlenecks))
		for i, b := range res.Bottlenecks {
			fmt.Printf("[%d] %s (%s) on %s\n", i+1, b.Type, b.Severity, b.NodeType)
			fmt.Printf("    Message: %s\n", b.Message)
			fmt.Printf("    Advice : %s\n", b.Recommendation)
		}
	} else {
		fmt.Println("\nNo severe bottlenecks detected in execution plan.")
	}
}

func runBloat(args []string) {
	fs := flag.NewFlagSet("bloat", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	demo := fs.Bool("demo", false, "Use demo data")
	_ = fs.Parse(args)

	var bloatList []model.BloatInfo
	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		bloatList, err = watchdog.EstimateTableBloat(ctx, client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bloat query error: %v\n", err)
			os.Exit(1)
		}
	} else {
		bloatList = model.GetDemoBloat()
	}

	fmt.Printf("\n%-25s %-12s %-12s %-12s %-10s\n", "Table", "Total Size", "Dead Tuples", "Bloat (Est)", "Dead %")
	fmt.Println(strings.Repeat("-", 75))
	for _, b := range bloatList {
		fmt.Printf("%-25s %-12s %-12d %-12s %-9.1f%%\n",
			b.TableName, b.SizePretty, b.DeadTuples, b.BloatPretty, b.DeadTupleRatio)
	}
}

func runAdvice(args []string) {
	fs := flag.NewFlagSet("advice", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	demo := fs.Bool("demo", false, "Use demo data")
	_ = fs.Parse(args)

	var recs []model.IndexRecommendation
	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		tables, err := client.GetTableStats(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read table stats: %v\n", err)
			os.Exit(1)
		}

		queries, err := client.GetSlowQueries(ctx, 100, "total")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read query stats: %v\n", err)
			os.Exit(1)
		}

		recs = synthesizer.AdviseIndexes(tables, queries)
	} else {
		recs = model.GetDemoRecommendations()
	}

	fmt.Printf("\n=== Missing Index Recommendations (%d) ===\n\n", len(recs))
	for i, r := range recs {
		fmt.Printf("[%d] Priority: %s | Table: %s | Est. Size: %s (%s write penalty)\n",
			i+1, r.Priority, r.TableName, r.EstimatedSizePretty, r.WritePenalty)
		fmt.Printf("    Reason : %s\n", r.Reason)
		fmt.Printf("    DDL    : %s\n\n", r.DDL)
	}
}

func runUnused(args []string) {
	fs := flag.NewFlagSet("unused", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	demo := fs.Bool("demo", false, "Use demo data")
	_ = fs.Parse(args)

	var unused []model.UnusedIndex
	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		unused, err = client.GetUnusedIndexes(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unused indexes query error: %v\n", err)
			os.Exit(1)
		}
	} else {
		unused = model.GetDemoUnusedIndexes()
	}

	fmt.Printf("\n=== Unused & Redundant Index Cleanup Candidates (%d) ===\n\n", len(unused))
	for i, u := range unused {
		status := "0 Scans"
		if u.IsRedundant {
			status = "Redundant"
		}
		fmt.Printf("[%d] %s on %s (%s, %s)\n", i+1, u.IndexName, u.TableName, u.IndexSizePretty, status)
		fmt.Printf("    Reason : %s\n", u.Reason)
		fmt.Printf("    Cleanup: %s\n\n", u.DropDDL)
	}
}

func runLocks(args []string) {
	fs := flag.NewFlagSet("locks", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	demo := fs.Bool("demo", false, "Use demo data")
	_ = fs.Parse(args)

	var tree model.LockContentionTree
	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		tree, err = client.GetLockContention(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Lock query error: %v\n", err)
			os.Exit(1)
		}
	} else {
		tree = model.GetDemoLockContention()
	}

	fmt.Printf("\n=== Live Activity & Lock Contention ===\n")
	fmt.Printf("Active Connections: %d | Blocked Queries: %d | Contention Detected: %v\n\n",
		len(tree.ActiveProcesses), tree.TotalBlockedQueries, tree.HasContention)

	if len(tree.RootBlockers) > 0 {
		fmt.Println("--- Root Blocking Queries ---")
		for _, b := range tree.RootBlockers {
			fmt.Printf("PID %d (User: %s, State: %s, Duration: %.1f ms)\n", b.PID, b.User, b.State, b.DurationMs)
			fmt.Printf("  Query: %s\n", b.Query)
			fmt.Printf("  Blocking PIDs: %v\n\n", b.BlockedPIDs)
		}
	} else {
		fmt.Println("No queries are currently blocked on table or row locks.")
	}
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	demo := fs.Bool("demo", false, "Use demo data")
	failBloat := fs.Float64("fail-on-bloat", 35.0, "Fail if any table dead tuple ratio exceeds this percentage")
	failCache := fs.Float64("fail-on-cache", 90.0, "Fail if overall buffer cache hit ratio drops below this percentage")
	failWrap := fs.Float64("fail-on-wraparound", 75.0, "Fail if TxID wraparound consumption exceeds this percentage")
	_ = fs.Parse(args)

	var cacheStats model.CacheHitStats
	var bloatList []model.BloatInfo
	var wrap model.WraparoundInfo

	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		cacheStats, _ = client.GetCacheStats(ctx)
		bloatList, _ = watchdog.EstimateTableBloat(ctx, client)
		wrap, _ = watchdog.CheckWraparound(ctx, client)
	} else {
		cacheStats = model.GetDemoCacheStats()
		bloatList = model.GetDemoBloat()
		_, wrap = model.GetDemoAutovacuum()
	}

	failed := false
	fmt.Println("=== IndexLens Health Quality Gate ===")

	// 1. Check cache hit ratio
	fmt.Printf("[Check] Buffer Cache Hit Ratio: %.1f%% (Threshold: >= %.1f%%)... ",
		cacheStats.OverallHitRatio, *failCache)
	if cacheStats.OverallHitRatio < *failCache {
		fmt.Println("FAILED")
		failed = true
	} else {
		fmt.Println("PASSED")
	}

	// 2. Check table bloat
	maxDeadRatio := 0.0
	worstTable := ""
	for _, b := range bloatList {
		if b.DeadTupleRatio > maxDeadRatio {
			maxDeadRatio = b.DeadTupleRatio
			worstTable = b.TableName
		}
	}
	fmt.Printf("[Check] Max Table Dead Tuples: %.1f%% on '%s' (Threshold: <= %.1f%%)... ",
		maxDeadRatio, worstTable, *failBloat)
	if maxDeadRatio > *failBloat {
		fmt.Println("FAILED")
		failed = true
	} else {
		fmt.Println("PASSED")
	}

	// 3. Check wraparound
	fmt.Printf("[Check] TxID Wraparound Consumed: %.1f%% (Threshold: <= %.1f%%)... ",
		wrap.PercentConsumed, *failWrap)
	if wrap.PercentConsumed > *failWrap {
		fmt.Println("FAILED")
		failed = true
	} else {
		fmt.Println("PASSED")
	}

	if failed {
		fmt.Println("\nQuality Gate FAILED: Performance or safety thresholds violated.")
		os.Exit(1)
	}

	fmt.Println("\nQuality Gate PASSED: Database metrics meet all reliability standards.")
}

func runReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	dsn := fs.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	output := fs.String("output", "indexlens_report.md", "Output file path for Markdown report")
	demo := fs.Bool("demo", false, "Use demo data")
	_ = fs.Parse(args)

	var queries []model.QueryStat
	var recs []model.IndexRecommendation
	var unused []model.UnusedIndex
	var bloat []model.BloatInfo

	if !*demo && *dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := poller.NewClient(ctx, *dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()

		queries, _ = client.GetSlowQueries(ctx, 10, "total")
		tables, _ := client.GetTableStats(ctx)
		recs = synthesizer.AdviseIndexes(tables, queries)
		unused, _ = client.GetUnusedIndexes(ctx)
		bloat, _ = watchdog.EstimateTableBloat(ctx, client)
	} else {
		queries = model.GetDemoQueries()
		recs = model.GetDemoRecommendations()
		unused = model.GetDemoUnusedIndexes()
		bloat = model.GetDemoBloat()
	}

	var md strings.Builder
	md.WriteString("# IndexLens PostgreSQL Performance & Diagnostic Report\n\n")
	md.WriteString(fmt.Sprintf("*Generated on: %s*\n\n", time.Now().UTC().Format(time.RFC1123)))

	md.WriteString("## 1. Top Identified Missing Indexes\n\n")
	for i, r := range recs {
		md.WriteString(fmt.Sprintf("### [%d] Table: `%s` (Priority: %s)\n", i+1, r.TableName, r.Priority))
		md.WriteString(fmt.Sprintf("- **Reason**: %s\n", r.Reason))
		md.WriteString(fmt.Sprintf("- **Estimated Size**: %s | **Write Penalty**: %s\n", r.EstimatedSizePretty, r.WritePenalty))
		md.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", r.DDL))
	}

	if len(unused) > 0 {
		md.WriteString("## 2. Unused & Redundant Indexes (Cleanup Candidates)\n\n")
		for i, u := range unused {
			md.WriteString(fmt.Sprintf("### [%d] Index: `%s` on `%s` (%s)\n", i+1, u.IndexName, u.TableName, u.IndexSizePretty))
			md.WriteString(fmt.Sprintf("- **Reason**: %s\n", u.Reason))
			md.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", u.DropDDL))
		}
	}

	md.WriteString("## 3. High Latency Queries\n\n")
	md.WriteString("| Calls | Mean Latency | Total Duration | Cache Hit % | Query Snippet |\n")
	md.WriteString("|---|---|---|---|---|\n")
	for _, q := range queries {
		snippet := q.Query
		if len(snippet) > 60 {
			snippet = snippet[:57] + "..."
		}
		md.WriteString(fmt.Sprintf("| %d | %.1f ms | %.1f s | %.1f%% | `%s` |\n",
			q.Calls, q.MeanExecTime, q.TotalExecTime/1000.0, q.CacheHitRatio, strings.ReplaceAll(snippet, "|", "/")))
	}
	md.WriteString("\n")

	md.WriteString("## 4. Table Bloat & Dead Tuples\n\n")
	md.WriteString("| Relation | Total Size | Dead Tuples | Dead % | Estimated Bloat |\n")
	md.WriteString("|---|---|---|---|---|\n")
	for _, b := range bloat {
		md.WriteString(fmt.Sprintf("| `%s` | %s | %d | %.1f%% | %s |\n",
			b.TableName, b.SizePretty, b.DeadTuples, b.DeadTupleRatio, b.BloatPretty))
	}

	err := os.WriteFile(*output, []byte(md.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write report file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[IndexLens] Report generated successfully: %s\n", *output)
}

func maskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) > 1 {
		return "postgres://***@" + parts[1]
	}
	return "postgres://..."
}

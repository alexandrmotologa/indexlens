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

const version = "1.0.0"

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

Commands:
  serve     Start the interactive web studio and REST API
  explain   Parse and diagnose an EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) plan
  bloat     Estimate table dead tuples and disk bloat
  advice    Synthesize non-blocking CREATE INDEX CONCURRENTLY recommendations
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
		// Output demo plan diagnostic
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
		fmt.Printf("[%d] Priority: %s | Table: %s\n", i+1, r.Priority, r.TableName)
		fmt.Printf("    Reason : %s\n", r.Reason)
		fmt.Printf("    DDL    : %s\n\n", r.DDL)
	}
}

func maskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) > 1 {
		return "postgres://***@" + parts[1]
	}
	return "postgres://..."
}

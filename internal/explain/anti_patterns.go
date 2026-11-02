package explain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

var (
	fnCallRegex       = regexp.MustCompile(`(?i)\b(lower|upper|date|date_trunc|substring|trim)\s*\(\s*([a-zA-Z0-9_\.]+)\s*\)\s*(=|!=|<>|LIKE\b|ILIKE\b|IN\b)`)
	leadingWildcardRe = regexp.MustCompile(`(?i)\b([a-zA-Z0-9_\.]+)\s+(LIKE|ILIKE)\s*'(%[^'\n]+)'`)
	deepOffsetRegex   = regexp.MustCompile(`(?i)\bOFFSET\s+(\d+)\b`)
	selectStarRegex   = regexp.MustCompile(`(?i)\bSELECT\s+\*\s+FROM\b`)
)

// AnalyzeSQLAntiPatterns inspects a SQL query string and returns flagged anti-patterns.
func AnalyzeSQLAntiPatterns(query string) []model.SQLAntiPattern {
	var antiPatterns []model.SQLAntiPattern
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return antiPatterns
	}

	// 1. Function Call on Column (e.g. lower(email) = ...)
	if matches := fnCallRegex.FindAllStringSubmatch(cleanQuery, -1); len(matches) > 0 {
		for _, m := range matches {
			fnName := strings.ToLower(m[1])
			colExpr := m[2]
			op := m[3]

			parts := strings.Split(colExpr, ".")
			colName := parts[len(parts)-1]
			tableName := "target_table"
			if len(parts) > 1 {
				tableName = parts[len(parts)-2]
			}

			antiPatterns = append(antiPatterns, model.SQLAntiPattern{
				Type:     "FUNCTION_ON_INDEX_COLUMN",
				Severity: "CRITICAL",
				Snippet:  fmt.Sprintf("%s(%s) %s", fnName, colExpr, op),
				Message:  fmt.Sprintf("Calling function '%s' on column '%s' invalidates standard B-Tree index lookups and causes full sequential scans.", fnName, colExpr),
				RewriteAdvice: fmt.Sprintf("Either store normalized values in '%s' or create an expression index to accelerate this filter.", colName),
				SuggestedDDL:  fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expr_%s_%s ON %s (%s(%s));", tableName, colName, tableName, fnName, colName),
			})
		}
	}

	// 2. Leading Wildcard in LIKE/ILIKE (e.g. LIKE '%term%')
	if matches := leadingWildcardRe.FindAllStringSubmatch(cleanQuery, -1); len(matches) > 0 {
		for _, m := range matches {
			colExpr := m[1]
			operator := strings.ToUpper(m[2])
			pattern := m[3]

			parts := strings.Split(colExpr, ".")
			colName := parts[len(parts)-1]
			tableName := "target_table"
			if len(parts) > 1 {
				tableName = parts[len(parts)-2]
			}

			antiPatterns = append(antiPatterns, model.SQLAntiPattern{
				Type:     "LEADING_WILDCARD",
				Severity: "CRITICAL",
				Snippet:  fmt.Sprintf("%s %s '%s'", colExpr, operator, pattern),
				Message:  fmt.Sprintf("Leading wildcard pattern '%s' prevents B-Tree bisection, forcing table scans or expensive full index scans.", pattern),
				RewriteAdvice: "For substring and suffix matching, utilize PostgreSQL's pg_trgm extension with a GIN index instead of plain LIKE.",
				SuggestedDDL: fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS pg_trgm;\nCREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gin_%s_%s ON %s USING gin (%s gin_trgm_ops);",
					tableName, colName, tableName, colName),
			})
		}
	}

	// 3. Deep OFFSET pagination (e.g. OFFSET 25000)
	if matches := deepOffsetRegex.FindAllStringSubmatch(cleanQuery, -1); len(matches) > 0 {
		for _, m := range matches {
			offsetVal, err := strconv.ParseInt(m[1], 10, 64)
			if err == nil && offsetVal >= 500 {
				antiPatterns = append(antiPatterns, model.SQLAntiPattern{
					Type:     "DEEP_OFFSET",
					Severity: "WARNING",
					Snippet:  fmt.Sprintf("OFFSET %d", offsetVal),
					Message:  fmt.Sprintf("Deep OFFSET of %d rows forces PostgreSQL to read and discard all prior rows, causing severe buffer churn and disk IO.", offsetVal),
					RewriteAdvice: "Switch from numeric OFFSET to Keyset (Seek) Pagination using indexed cursor columns (e.g. WHERE id > :last_seen_id ORDER BY id ASC LIMIT :page_size).",
				})
			}
		}
	}

	// 4. SELECT * on complex or multi-table queries
	if selectStarRegex.MatchString(cleanQuery) && strings.Contains(strings.ToUpper(cleanQuery), "JOIN") {
		antiPatterns = append(antiPatterns, model.SQLAntiPattern{
			Type:          "SELECT_STAR",
			Severity:      "SUGGESTION",
			Snippet:       "SELECT * FROM",
			Message:       "Wildcard SELECT on joined tables inflates network serialization overhead and prevents PostgreSQL from utilizing Index-Only Scans.",
			RewriteAdvice: "Explicitly enumerate required columns to enable index-only covering paths and conserve buffer cache.",
		})
	}

	return antiPatterns
}

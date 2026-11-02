package synthesizer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

var (
	// Regex patterns for extracting clauses from SQL strings
	whereClauseRegex = regexp.MustCompile(`(?i)\bWHERE\b\s+(.+?)(?:\bORDER\b|\bGROUP\b|\bHAVING\b|\bLIMIT\b|$)`)
	orderByRegex     = regexp.MustCompile(`(?i)\bORDER\s+BY\b\s+(.+?)(?:\bLIMIT\b|$)`)
	fromTableRegex   = regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z0-9_\.]+)(?:\s+(?:AS\s+)?([a-zA-Z0-9_]+))?`)
	conditionRegex   = regexp.MustCompile(`([a-zA-Z0-9_\.]+)\s*(=|!=|<>|<=|>=|<|>|\bBETWEEN\b|\bIN\b|\bLIKE\b|\bIS\s+NULL\b|\bIS\s+NOT\b)`)
)

// AdviseIndexes examines table stats and slow queries to generate missing index recommendations.
func AdviseIndexes(tables []model.TableStat, queries []model.QueryStat) []model.IndexRecommendation {
	var recommendations []model.IndexRecommendation

	// 1. Identify tables under sequential scan pressure
	seqScanTables := make(map[string]model.TableStat)
	for _, t := range tables {
		totalScans := t.SeqScan + t.IdxScan
		// Flag if > 25 sequential scans and either seq > idx or high tuple read
		if t.SeqScan >= 25 && (float64(t.SeqScan)/float64(totalScans+1) > 0.3 || t.SeqTupRead > 100000) {
			seqScanTables[strings.ToLower(t.TableName)] = t
		}
	}

	// 2. Correlate with slow queries targeting these tables
	recIndex := 1
	handledTables := make(map[string]bool)

	for _, q := range queries {
		targetTable := extractTargetTable(q.Query)
		if targetTable == "" {
			continue
		}

		tblStat, isPressure := seqScanTables[strings.ToLower(targetTable)]
		if !isPressure || handledTables[strings.ToLower(targetTable)] {
			continue
		}

		candidates, partialPredicate := extractCandidatesFromSQL(q.Query)
		if len(candidates) == 0 {
			continue
		}

		rankedCols := RankColumns(candidates)
		if len(rankedCols) == 0 {
			continue
		}

		indexName := GenerateIndexName(targetTable, rankedCols)
		ddl := GenerateDDL(tblStat.SchemaName, targetTable, indexName, rankedCols, partialPredicate)

		priority := "MEDIUM"
		if tblStat.SeqScan > 1000 || tblStat.SeqTupRead > 1000000 || q.MeanExecTime > 50.0 {
			priority = "HIGH"
		}

		// Estimate scan reduction percentage based on number of columns and selectivity
		estReduction := 85.0
		if len(rankedCols) >= 2 {
			estReduction = 92.5
		}

		recID := fmt.Sprintf("rec-%s-%d", targetTable, recIndex)
		recIndex++

		reason := fmt.Sprintf("%d sequential scans reading %s rows on '%s'. Query filters on %s with mean latency %.1f ms.",
			tblStat.SeqScan, model.FormatBytes(tblStat.SeqTupRead), targetTable, strings.Join(rankedCols, ", "), q.MeanExecTime)

		estBytes, estPretty, writePenalty := EstimateIndexSize(rankedCols, tblStat.LiveTuples)

		recommendations = append(recommendations, model.IndexRecommendation{
			ID:                     recID,
			TableName:              targetTable,
			IndexName:              indexName,
			Columns:                rankedCols,
			Predicate:              partialPredicate,
			Reason:                 reason,
			DDL:                    ddl,
			EstimatedScanReduction: estReduction,
			EstimatedSizeBytes:     estBytes,
			EstimatedSizePretty:    estPretty,
			WritePenalty:           writePenalty,
			SeqScanCount:           tblStat.SeqScan,
			SeqTuplesRead:          tblStat.SeqTupRead,
			Priority:               priority,
		})

		handledTables[strings.ToLower(targetTable)] = true
	}

	return recommendations
}

func extractTargetTable(sqlQuery string) string {
	matches := fromTableRegex.FindStringSubmatch(sqlQuery)
	if len(matches) > 1 {
		tbl := matches[1]
		parts := strings.Split(tbl, ".")
		return parts[len(parts)-1]
	}
	return ""
}

func extractCandidatesFromSQL(sqlQuery string) ([]ColumnCandidate, string) {
	var candidates []ColumnCandidate
	var partialPredicate string

	// Extract WHERE clause
	whereMatches := whereClauseRegex.FindStringSubmatch(sqlQuery)
	if len(whereMatches) > 1 {
		whereText := whereMatches[1]
		condMatches := conditionRegex.FindAllStringSubmatch(whereText, -1)

		for _, m := range condMatches {
			colRaw := m[1]
			op := strings.ToUpper(strings.TrimSpace(m[2]))

			// Strip table aliases like "u.tenant_id" -> "tenant_id"
			colParts := strings.Split(colRaw, ".")
			colName := colParts[len(colParts)-1]

			if strings.EqualFold(colName, "deleted_at") && strings.Contains(op, "IS NULL") {
				partialPredicate = "deleted_at IS NULL"
				continue
			}

			if strings.EqualFold(colName, "processed") && strings.Contains(whereText, "processed = false") {
				partialPredicate = "processed = false"
				continue
			}

			if op == "=" || strings.Contains(op, "IS NULL") {
				candidates = append(candidates, ColumnCandidate{
					ColumnName: colName,
					Type:       PredEquality,
				})
			} else {
				candidates = append(candidates, ColumnCandidate{
					ColumnName: colName,
					Type:       PredRange,
				})
			}
		}
	}

	// Extract ORDER BY clause
	orderMatches := orderByRegex.FindStringSubmatch(sqlQuery)
	if len(orderMatches) > 1 {
		orderText := orderMatches[1]
		items := strings.Split(orderText, ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			direction := "ASC"
			if strings.Contains(strings.ToUpper(item), "DESC") {
				direction = "DESC"
			}

			tokens := strings.Fields(item)
			if len(tokens) > 0 {
				colParts := strings.Split(tokens[0], ".")
				colName := colParts[len(colParts)-1]
				candidates = append(candidates, ColumnCandidate{
					ColumnName: colName,
					Type:       PredSort,
					Direction:  direction,
				})
			}
		}
	}

	return candidates, partialPredicate
}

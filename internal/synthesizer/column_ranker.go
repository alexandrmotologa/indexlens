package synthesizer

import (
	"strings"
)

// PredicateType indicates how a column is filtered in a query.
type PredicateType int

const (
	PredEquality PredicateType = iota // =, IS NULL
	PredRange                         // <, >, <=, >=, BETWEEN, IN, LIKE
	PredSort                          // ORDER BY
)

// ColumnCandidate represents a candidate column for indexing with its role.
type ColumnCandidate struct {
	ColumnName string
	Type       PredicateType
	Direction  string // "ASC", "DESC", or ""
}

// RankColumns orders columns based on standard B-Tree indexing rules:
// 1. Equality predicates first
// 2. Inequality/range predicates second
// 3. Sort keys last
func RankColumns(candidates []ColumnCandidate) []string {
	var equalityCols []string
	var rangeCols []string
	var sortCols []string

	seen := make(map[string]bool)

	// First pass: group by predicate type
	for _, c := range candidates {
		colClean := strings.TrimSpace(c.ColumnName)
		if colClean == "" {
			continue
		}
		baseCol := strings.ToLower(strings.Fields(colClean)[0])
		if seen[baseCol] {
			continue
		}

		switch c.Type {
		case PredEquality:
			equalityCols = append(equalityCols, colClean)
			seen[baseCol] = true
		case PredRange:
			rangeCols = append(rangeCols, colClean)
			seen[baseCol] = true
		case PredSort:
			sortFormatted := colClean
			if c.Direction != "" && !strings.Contains(strings.ToUpper(colClean), "ASC") && !strings.Contains(strings.ToUpper(colClean), "DESC") {
				sortFormatted = colClean + " " + c.Direction
			}
			sortCols = append(sortCols, sortFormatted)
			seen[baseCol] = true
		}
	}

	// Combine in standard order: Equality -> Range -> Sort
	var ordered []string
	ordered = append(ordered, equalityCols...)
	ordered = append(ordered, rangeCols...)
	ordered = append(ordered, sortCols...)

	return ordered
}

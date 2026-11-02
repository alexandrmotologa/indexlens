package synthesizer

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var nonAlphaNumRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// GenerateIndexName produces a clean, deterministic PostgreSQL identifier under 63 bytes.
func GenerateIndexName(tableName string, columns []string) string {
	var cleanCols []string
	for _, col := range columns {
		c := strings.Fields(col)[0] // strip ASC/DESC
		c = nonAlphaNumRegex.ReplaceAllString(c, "")
		if c != "" {
			cleanCols = append(cleanCols, c)
		}
	}

	rawName := fmt.Sprintf("idx_%s_%s", tableName, strings.Join(cleanCols, "_"))
	if len(rawName) <= 60 {
		return rawName
	}

	// Truncate and append short hash to stay strictly within PostgreSQL's 63-byte NAMEDATALEN limit
	hasher := md5.Sum([]byte(rawName))
	hashSuffix := hex.EncodeToString(hasher[:])[:6]
	truncated := rawName[:53]
	return fmt.Sprintf("%s_%s", truncated, hashSuffix)
}

// GenerateDDL constructs the non-blocking CREATE INDEX CONCURRENTLY statement.
func GenerateDDL(schema, tableName, indexName string, columns []string, predicate string) string {
	if schema == "" {
		schema = "public"
	}

	colsJoined := strings.Join(columns, ", ")
	var whereClause string
	if predicate != "" {
		whereClause = fmt.Sprintf(" WHERE %s", strings.TrimSpace(predicate))
	}

	return fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s.%s (%s)%s;",
		indexName, schema, tableName, colsJoined, whereClause)
}

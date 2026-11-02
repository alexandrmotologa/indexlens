package poller

import (
	"context"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/indexlens/internal/model"
)

// GetSchemaHealthReport runs non-invasive schema inspections to detect unindexed foreign keys,
// tables missing primary keys, and relations suffering from heavy TOAST bloat.
func (c *Client) GetSchemaHealthReport(ctx context.Context) (model.SchemaHealthReport, error) {
	report := model.SchemaHealthReport{
		UnindexedFKs:   make([]model.UnindexedForeignKey, 0),
		MissingPKs:     make([]model.TableMissingPK, 0),
		ToastRelations: make([]model.ToastRelationInfo, 0),
	}

	// 1. Unindexed Foreign Keys
	// Checks foreign keys in pg_constraint against pg_index to detect missing supporting indexes
	fkQuery := `
		SELECT
			c.conname AS constraint_name,
			child_tbl.relname AS child_table,
			parent_tbl.relname AS parent_table,
			ARRAY_AGG(att.attname ORDER BY u.attpos) AS fk_columns
		FROM pg_constraint c
		JOIN pg_class child_tbl ON c.conrelid = child_tbl.oid
		JOIN pg_namespace ns ON child_tbl.relnamespace = ns.oid
		JOIN pg_class parent_tbl ON c.confrelid = parent_tbl.oid
		CROSS JOIN LATERAL UNNEST(c.conkey) WITH ORDINALITY AS u(attnum, attpos)
		JOIN pg_attribute att ON att.attrelid = c.conrelid AND att.attnum = u.attnum
		WHERE c.contype = 'f'
		  AND ns.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND NOT EXISTS (
			  SELECT 1
			  FROM pg_index i
			  WHERE i.indrelid = c.conrelid
			    AND (i.indkey::smallint[])[0:cardinality(c.conkey)-1] = c.conkey
		  )
		GROUP BY c.conname, child_tbl.relname, parent_tbl.relname
		ORDER BY child_tbl.relname, c.conname
	`

	fkRows, err := c.pool.Query(ctx, fkQuery)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var fk model.UnindexedForeignKey
			var cols []string
			if err := fkRows.Scan(&fk.ConstraintName, &fk.ChildTable, &fk.ParentTable, &cols); err == nil {
				fk.Columns = cols
				colStr := strings.Join(cols, ", ")
				colSuffix := strings.Join(cols, "_")
				if len(colSuffix) > 30 {
					colSuffix = colSuffix[:30]
				}
				idxName := fmt.Sprintf("idx_fk_%s_%s", fk.ChildTable, colSuffix)
				if len(idxName) > 63 {
					idxName = idxName[:63]
				}
				fk.SuggestedDDL = fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS %s ON %s (%s);", idxName, fk.ChildTable, colStr)
				fk.Severity = "CRITICAL"
				fk.Reason = fmt.Sprintf("Lacks supporting index on child table '%s'. Modifications or deletes on parent table '%s' will acquire full table locks or trigger sequential scans.", fk.ChildTable, fk.ParentTable)
				report.UnindexedFKs = append(report.UnindexedFKs, fk)
			}
		}
	}

	// 2. Tables Missing Primary Keys
	pkQuery := `
		SELECT
			ns.nspname AS schema_name,
			c.relname AS table_name,
			c.reltuples::bigint AS estimated_rows,
			pg_size_pretty(pg_total_relation_size(c.oid)) AS total_size
		FROM pg_class c
		JOIN pg_namespace ns ON c.relnamespace = ns.oid
		WHERE c.relkind = 'r'
		  AND ns.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND NOT EXISTS (
			  SELECT 1
			  FROM pg_constraint con
			  WHERE con.conrelid = c.oid
			    AND con.contype = 'p'
		  )
		ORDER BY c.reltuples DESC
		LIMIT 20
	`

	pkRows, err := c.pool.Query(ctx, pkQuery)
	if err == nil {
		defer pkRows.Close()
		for pkRows.Next() {
			var mpk model.TableMissingPK
			if err := pkRows.Scan(&mpk.SchemaName, &mpk.TableName, &mpk.EstimatedRows, &mpk.TotalSizePretty); err == nil {
				mpk.SuggestedDDL = fmt.Sprintf("ALTER TABLE %s.%s ADD PRIMARY KEY (id);", mpk.SchemaName, mpk.TableName)
				report.MissingPKs = append(report.MissingPKs, mpk)
			}
		}
	}

	// 3. TOAST Relation Storage Analysis
	toastQuery := `
		SELECT
			c.relname AS table_name,
			t.relname AS toast_table_name,
			pg_relation_size(c.oid) AS main_size,
			pg_total_relation_size(c.reltoastrelid) AS toast_size
		FROM pg_class c
		JOIN pg_class t ON c.reltoastrelid = t.oid
		JOIN pg_namespace ns ON c.relnamespace = ns.oid
		WHERE c.relkind = 'r'
		  AND c.reltoastrelid != 0
		  AND ns.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND pg_total_relation_size(c.reltoastrelid) > 10485760 -- > 10MB
		ORDER BY toast_size DESC
		LIMIT 20
	`

	toastRows, err := c.pool.Query(ctx, toastQuery)
	if err == nil {
		defer toastRows.Close()
		for toastRows.Next() {
			var tInfo model.ToastRelationInfo
			if err := toastRows.Scan(&tInfo.TableName, &tInfo.ToastTableName, &tInfo.MainSizeBytes, &tInfo.ToastSizeBytes); err == nil {
				total := tInfo.MainSizeBytes + tInfo.ToastSizeBytes
				if total > 0 {
					tInfo.ToastRatio = (float64(tInfo.ToastSizeBytes) / float64(total)) * 100.0
				}
				tInfo.MainSizePretty = model.FormatBytes(tInfo.MainSizeBytes)
				tInfo.ToastSizePretty = model.FormatBytes(tInfo.ToastSizeBytes)
				if tInfo.ToastRatio > 60.0 {
					tInfo.Recommendation = "High out-of-line storage ratio. Consider enabling LZ4 column compression (ALTER TABLE ... SET COMPRESSION lz4) or offloading binary/JSON blobs."
				} else {
					tInfo.Recommendation = "Moderate out-of-line storage. Monitor row width and ensure wide columns are only queried when strictly required."
				}
				report.ToastRelations = append(report.ToastRelations, tInfo)
			}
		}
	}

	report.TotalIssues = len(report.UnindexedFKs) + len(report.MissingPKs)
	// Base health score calculation out of 100
	score := 100 - (len(report.UnindexedFKs) * 15) - (len(report.MissingPKs) * 10)
	if score < 0 {
		score = 0
	}
	report.HealthScore = score

	return report, nil
}

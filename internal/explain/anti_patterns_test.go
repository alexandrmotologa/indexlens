package explain

import (
	"testing"
)

func TestAnalyzeSQLAntiPatterns(t *testing.T) {
	t.Run("Function on indexed column", func(t *testing.T) {
		q := "SELECT id, email FROM users WHERE lower(email) = 'test@example.com'"
		patterns := AnalyzeSQLAntiPatterns(q)
		if len(patterns) == 0 {
			t.Fatalf("expected anti-pattern for lower(email), got none")
		}
		found := false
		for _, p := range patterns {
			if p.Type == "FUNCTION_ON_INDEX_COLUMN" {
				found = true
				if p.SuggestedDDL == "" {
					t.Errorf("expected suggested DDL, got empty string")
				}
			}
		}
		if !found {
			t.Errorf("expected FUNCTION_ON_INDEX_COLUMN anti-pattern")
		}
	})

	t.Run("Leading wildcard LIKE", func(t *testing.T) {
		q := "SELECT id, name FROM products WHERE name ILIKE '%phone%'"
		patterns := AnalyzeSQLAntiPatterns(q)
		found := false
		for _, p := range patterns {
			if p.Type == "LEADING_WILDCARD" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected LEADING_WILDCARD anti-pattern")
		}
	})

	t.Run("Deep offset pagination", func(t *testing.T) {
		q := "SELECT id, title FROM articles ORDER BY published_at DESC LIMIT 50 OFFSET 10000"
		patterns := AnalyzeSQLAntiPatterns(q)
		found := false
		for _, p := range patterns {
			if p.Type == "DEEP_OFFSET" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected DEEP_OFFSET anti-pattern")
		}
	})

	t.Run("Clean query with normal indexable predicates", func(t *testing.T) {
		q := "SELECT id, status FROM orders WHERE user_id = 42 AND status = 'COMPLETED' LIMIT 10"
		patterns := AnalyzeSQLAntiPatterns(q)
		if len(patterns) != 0 {
			t.Errorf("expected 0 anti-patterns for clean query, got %d", len(patterns))
		}
	})
}

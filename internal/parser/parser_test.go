package parser

import "testing"

func TestParseDetectsPatterns(t *testing.T) {
	sql := `WITH x AS (SELECT * FROM users WHERE payload->>'name' = 'budi') SELECT DISTINCT ON (id) * FROM x ORDER BY id, created_at DESC LIMIT 10 OFFSET 20`
	qi := Parse(sql)
	if !qi.HasCTE || !qi.HasJSONB || !qi.HasDistinctOn || !qi.HasOrderBy || !qi.HasOffset {
		t.Fatalf("expected patterns to be detected: %+v", qi)
	}
}

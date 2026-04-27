package tests

import (
	"strings"
	"testing"

	"github.com/budisuryadi/pgquerydoctor/internal/analyzer"
	"github.com/budisuryadi/pgquerydoctor/internal/report"
)

func TestAnalyzerProducesReport(t *testing.T) {
	sql := `SELECT * FROM users WHERE email = 'a@b.com' ORDER BY created_at DESC LIMIT 10 OFFSET 20`
	plan := `Seq Scan on users (cost=0.00..100.00 rows=200000 width=10) (actual time=1.00..2.00 rows=200000 loops=1)
Sort Method: external merge Disk: 1024kB
Planning Time: 1.000 ms
Execution Time: 50.000 ms`
	result := analyzer.Analyze(sql, plan)
	md := report.Markdown(result)
	if len(result.Findings) == 0 || !strings.Contains(md, "PgQueryDoctor Report") {
		t.Fatalf("expected findings and markdown report")
	}
}

package report

import (
	"fmt"
	"strings"

	"github.com/budisuryadi/pgquerydoctor/internal/analyzer"
)

func Markdown(r analyzer.Result) string {
	var b strings.Builder
	b.WriteString("# PgQueryDoctor Report\n\n")
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- Tables detected: `%s`\n", strings.Join(r.Query.Tables, "`, `")))
	b.WriteString(fmt.Sprintf("- Planning time: `%.3f ms`\n", r.Plan.PlanningTimeMS))
	b.WriteString(fmt.Sprintf("- Execution time: `%.3f ms`\n", r.Plan.ExecutionTimeMS))
	b.WriteString(fmt.Sprintf("- Findings: `%d`\n\n", len(r.Findings)))

	b.WriteString("## Query Shape\n\n")
	writeBool(&b, "SELECT", r.Query.HasSelect)
	writeBool(&b, "JOIN", r.Query.HasJoin)
	writeBool(&b, "WHERE", r.Query.HasWhere)
	writeBool(&b, "GROUP BY", r.Query.HasGroupBy)
	writeBool(&b, "ORDER BY", r.Query.HasOrderBy)
	writeBool(&b, "LIMIT", r.Query.HasLimit)
	writeBool(&b, "OFFSET", r.Query.HasOffset)
	writeBool(&b, "CTE", r.Query.HasCTE)
	writeBool(&b, "Subquery", r.Query.HasSubquery)
	writeBool(&b, "DISTINCT ON", r.Query.HasDistinctOn)
	writeBool(&b, "JSONB", r.Query.HasJSONB)
	b.WriteString("\n")

	b.WriteString("## Bottlenecks and Recommendations\n\n")
	for i, f := range r.Findings {
		b.WriteString(fmt.Sprintf("### %d. %s [%s]\n\n", i+1, f.Title, f.Severity))
		section(&b, "Rule", f.RuleID)
		section(&b, "Evidence", f.Evidence)
		section(&b, "Likely Root Cause", f.RootCause)
		section(&b, "Recommended Fix", f.Recommendation)
		section(&b, "Risk / Tradeoff", f.Tradeoff)
		if len(f.SuggestedIndexes) > 0 {
			b.WriteString("**Suggested Indexes**\n\n```sql\n")
			for _, idx := range f.SuggestedIndexes {
				b.WriteString(idx + "\n")
			}
			b.WriteString("```\n\n")
		}
		if f.OptimizedSQL != "" {
			b.WriteString("**Optimized SQL Example**\n\n```sql\n" + f.OptimizedSQL + "\n```\n\n")
		}
	}

	b.WriteString("## Next EXPLAIN Command\n\n")
	b.WriteString("```sql\nEXPLAIN (ANALYZE, BUFFERS, VERBOSE)\n")
	b.WriteString(strings.TrimSpace(r.Query.RawSQL))
	b.WriteString("\n```\n")
	return b.String()
}

func writeBool(b *strings.Builder, label string, v bool) {
	b.WriteString(fmt.Sprintf("- %s: `%t`\n", label, v))
}
func section(b *strings.Builder, title, value string) {
	if strings.TrimSpace(value) != "" {
		b.WriteString(fmt.Sprintf("**%s:** %s\n\n", title, value))
	}
}

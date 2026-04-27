package rules

import (
	"fmt"
	"strings"

	"github.com/budisuryadi/pgquerydoctor/internal/explain"
	"github.com/budisuryadi/pgquerydoctor/internal/parser"
)

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarn     Severity = "WARN"
	SeverityCritical Severity = "CRITICAL"
)

type Finding struct {
	RuleID           string
	Title            string
	Severity         Severity
	Evidence         string
	RootCause        string
	Recommendation   string
	Tradeoff         string
	SuggestedIndexes []string
	OptimizedSQL     string
}

type Context struct {
	Query parser.QueryInfo
	Plan  explain.PlanInfo
}

type Rule func(Context) []Finding

func Run(ctx Context) []Finding {
	rules := []Rule{seqScanRule, externalSortRule, nestedLoopRule, jsonbRule, orderLimitRule, offsetRule, distinctOnRule, havingMaxCaseRule, compositeWhereOrderRule, cteRule}
	out := []Finding{}
	for _, r := range rules {
		out = append(out, r(ctx)...)
	}
	if len(out) == 0 {
		out = append(out, Finding{RuleID: "OK001", Title: "No obvious high-risk pattern detected", Severity: SeverityInfo, Evidence: "PgQueryDoctor rule engine did not find common anti-patterns.", Recommendation: "Run EXPLAIN (ANALYZE, BUFFERS) in production-like data and compare with real latency metrics."})
	}
	return out
}

func seqScanRule(ctx Context) []Finding {
	if !ctx.Plan.HasSeqScan {
		return nil
	}
	tables := ctx.Plan.SeqScanTables
	idx := []string{}
	for _, t := range fallbackTables(tables, ctx.Query.Tables) {
		if len(ctx.Query.WhereColumns) > 0 {
			idx = append(idx, fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_filter ON %s (%s);", safeName(t), t, strings.Join(cleanCols(ctx.Query.WhereColumns), ", ")))
		}
	}
	return []Finding{{RuleID: "PG001", Title: "Sequential scan detected", Severity: SeverityWarn, Evidence: "EXPLAIN contains Seq Scan on: " + strings.Join(tables, ", "), RootCause: "PostgreSQL may be reading many rows because the filter is not selective or no suitable index exists.", Recommendation: "Check table cardinality, filter selectivity, and add an index only when the predicate is frequently used and selective.", Tradeoff: "Indexes speed reads but slow writes and increase storage.", SuggestedIndexes: idx}}
}

func externalSortRule(ctx Context) []Finding {
	if !ctx.Plan.HasExternalMerge {
		return nil
	}
	return []Finding{{RuleID: "PG002", Title: "Sort spilled to disk", Severity: SeverityCritical, Evidence: "EXPLAIN contains external merge / Disk / Temp I/O.", RootCause: "Sort could not fit in memory or ORDER BY/GROUP BY requires sorting many rows.", Recommendation: "Add a matching index for ORDER BY when possible, reduce rows earlier, or tune work_mem carefully per query/session.", Tradeoff: "Increasing work_mem globally can increase memory pressure under concurrency.", SuggestedIndexes: indexForOrder(ctx)}}
}

func nestedLoopRule(ctx Context) []Finding {
	if !ctx.Plan.HasNestedLoop {
		return nil
	}
	severity := SeverityWarn
	if len(ctx.Plan.HighRowNodes) > 0 {
		severity = SeverityCritical
	}
	return []Finding{{RuleID: "PG003", Title: "Nested Loop detected", Severity: severity, Evidence: "EXPLAIN contains Nested Loop.", RootCause: "Nested Loop can be expensive when the outer side returns many rows or the inner side lacks an index.", Recommendation: "Verify actual rows vs estimated rows. Add indexes on join keys, update statistics, or rewrite the query to reduce rows before join.", Tradeoff: "Hash Join can be faster for large inputs but may need more memory."}}
}

func jsonbRule(ctx Context) []Finding {
	if !ctx.Query.HasJSONB {
		return nil
	}
	idx := []string{}
	for _, t := range ctx.Query.Tables {
		idx = append(idx, fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_payload_gin ON %s USING GIN (payload jsonb_path_ops);", safeName(t), t))
	}
	return []Finding{{RuleID: "PG004", Title: "JSONB predicate detected", Severity: SeverityWarn, Evidence: "SQL contains JSONB operators/functions such as ->, @>, or jsonb_*.", RootCause: "JSONB filtering without a matching expression/GIN index can cause large scans.", Recommendation: "For containment (@>), consider GIN jsonb_path_ops. For equality on a specific path, consider an expression index.", Tradeoff: "GIN indexes can be large and slower to update.", SuggestedIndexes: idx}}
}

func orderLimitRule(ctx Context) []Finding {
	if !(ctx.Query.HasOrderBy && ctx.Query.HasLimit) {
		return nil
	}
	return []Finding{{RuleID: "PG005", Title: "ORDER BY + LIMIT pattern", Severity: SeverityInfo, Evidence: "SQL has ORDER BY and LIMIT.", RootCause: "Without a matching index, PostgreSQL may sort many rows before returning a small page.", Recommendation: "Create a composite index that starts with equality filters and ends with ORDER BY columns.", Tradeoff: "Composite index order matters; avoid creating many overlapping indexes.", SuggestedIndexes: indexForOrder(ctx)}}
}

func offsetRule(ctx Context) []Finding {
	if !ctx.Query.HasOffset {
		return nil
	}
	return []Finding{{RuleID: "PG006", Title: "OFFSET pagination anti-pattern", Severity: SeverityWarn, Evidence: "SQL contains OFFSET.", RootCause: "OFFSET forces PostgreSQL to skip rows, which gets slower on deep pages.", Recommendation: "Use keyset pagination, for example WHERE created_at < $last_created_at ORDER BY created_at DESC LIMIT $limit.", Tradeoff: "Keyset pagination needs stable ordering and different UI navigation behavior.", OptimizedSQL: "SELECT * FROM table_name WHERE created_at < $1 ORDER BY created_at DESC LIMIT $2;"}}
}

func distinctOnRule(ctx Context) []Finding {
	if !ctx.Query.HasDistinctOn {
		return nil
	}
	return []Finding{{RuleID: "PG007", Title: "DISTINCT ON requires careful index order", Severity: SeverityWarn, Evidence: "SQL contains DISTINCT ON.", RootCause: "DISTINCT ON is fast when index order matches DISTINCT ON keys followed by ORDER BY keys.", Recommendation: "Create an index that matches DISTINCT ON columns first, then the ORDER BY tiebreaker column.", Tradeoff: "Wrong column order may not help the planner."}}
}

func havingMaxCaseRule(ctx Context) []Finding {
	if !ctx.Query.HasHavingMaxCase {
		return nil
	}
	return []Finding{{RuleID: "PG008", Title: "GROUP BY + HAVING MAX(CASE WHEN...) pattern", Severity: SeverityCritical, Evidence: "SQL contains HAVING MAX(CASE WHEN ...).", RootCause: "This pattern often scans and aggregates many rows to emulate multi-attribute matching.", Recommendation: "Consider rewriting with EXISTS clauses, filtered partial indexes, or precomputed/search table if this is a hot path.", Tradeoff: "Rewrite can be longer but often allows better index usage."}}
}

func compositeWhereOrderRule(ctx Context) []Finding {
	if !(ctx.Query.HasWhere && ctx.Query.HasOrderBy) {
		return nil
	}
	return []Finding{{RuleID: "PG009", Title: "Potential composite index for WHERE + ORDER BY", Severity: SeverityInfo, Evidence: "SQL has both WHERE and ORDER BY.", RootCause: "Separate indexes may not avoid sorting or may not filter efficiently.", Recommendation: "For frequent queries, consider composite index: equality filters first, range filters next, ORDER BY last.", Tradeoff: "Do not add until verified with EXPLAIN (ANALYZE, BUFFERS).", SuggestedIndexes: indexForOrder(ctx)}}
}

func cteRule(ctx Context) []Finding {
	if !ctx.Query.HasCTE {
		return nil
	}
	return []Finding{{RuleID: "PG010", Title: "CTE detected", Severity: SeverityInfo, Evidence: "SQL starts with WITH.", RootCause: "CTEs can improve readability, but materialization/inlining behavior can affect plans depending on PostgreSQL version and query shape.", Recommendation: "Compare with subquery or NOT MATERIALIZED/MATERIALIZED when performance differs.", Tradeoff: "Readability and planner behavior should both be considered."}}
}

func indexForOrder(ctx Context) []string {
	if len(ctx.Query.Tables) == 0 || len(ctx.Query.OrderColumns) == 0 {
		return nil
	}
	table := ctx.Query.Tables[0]
	cols := append(cleanCols(ctx.Query.WhereColumns), cleanCols(ctx.Query.OrderColumns)...)
	if len(cols) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("CREATE INDEX CONCURRENTLY idx_%s_query_path ON %s (%s);", safeName(table), table, strings.Join(dedupe(cols), ", "))}
}
func fallbackTables(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}
func cleanCols(cols []string) []string {
	out := []string{}
	for _, c := range cols {
		c = strings.TrimSpace(strings.ReplaceAll(c, ";", ""))
		if c != "" && !strings.Contains(c, "->") {
			out = append(out, c)
		}
	}
	return out
}
func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
func safeName(s string) string { return strings.NewReplacer(".", "_", "-", "_").Replace(s) }

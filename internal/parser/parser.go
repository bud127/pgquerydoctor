package parser

import (
	"regexp"
	"strings"
)

type QueryInfo struct {
	RawSQL           string
	HasSelect        bool
	HasJoin          bool
	HasWhere         bool
	HasGroupBy       bool
	HasOrderBy       bool
	HasLimit         bool
	HasOffset        bool
	HasCTE           bool
	HasSubquery      bool
	HasDistinctOn    bool
	HasJSONB         bool
	HasHavingMaxCase bool
	Tables           []string
	WhereColumns     []string
	OrderColumns     []string
}

func Parse(sql string) QueryInfo {
	normalized := normalize(sql)
	qi := QueryInfo{RawSQL: sql}
	qi.HasSelect = strings.Contains(normalized, "select ")
	qi.HasJoin = strings.Contains(normalized, " join ")
	qi.HasWhere = strings.Contains(normalized, " where ")
	qi.HasGroupBy = strings.Contains(normalized, " group by ")
	qi.HasOrderBy = strings.Contains(normalized, " order by ")
	qi.HasLimit = strings.Contains(normalized, " limit ")
	qi.HasOffset = strings.Contains(normalized, " offset ")
	qi.HasCTE = strings.HasPrefix(strings.TrimSpace(normalized), "with ")
	qi.HasSubquery = regexp.MustCompile(`\(\s*select\s+`).MatchString(normalized)
	qi.HasDistinctOn = strings.Contains(normalized, "distinct on")
	qi.HasJSONB = strings.Contains(normalized, "->") || strings.Contains(normalized, "@>") || strings.Contains(normalized, "jsonb_")
	qi.HasHavingMaxCase = strings.Contains(normalized, "having") && strings.Contains(normalized, "max") && strings.Contains(normalized, "case when")
	qi.Tables = unique(extractTables(normalized))
	qi.WhereColumns = unique(extractColumnsAfter(normalized, "where", []string{"group by", "order by", "limit", "offset"}))
	qi.OrderColumns = unique(extractOrderColumns(normalized))
	return qi
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return " " + strings.TrimSpace(s) + " "
}

func extractTables(sql string) []string {
	re := regexp.MustCompile(`\b(from|join)\s+([a-zA-Z0-9_\.]+)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 2 && !strings.HasPrefix(m[2], "select") {
			out = append(out, m[2])
		}
	}
	return out
}

func extractColumnsAfter(sql, start string, stopWords []string) []string {
	idx := strings.Index(sql, " "+start+" ")
	if idx < 0 {
		return nil
	}
	part := sql[idx+len(start)+2:]
	end := len(part)
	for _, sw := range stopWords {
		if i := strings.Index(part, " "+sw+" "); i >= 0 && i < end {
			end = i
		}
	}
	part = part[:end]
	return extractColumnRefs(part)
}

func extractOrderColumns(sql string) []string {
	idx := strings.Index(sql, " order by ")
	if idx < 0 {
		return nil
	}
	part := sql[idx+10:]
	end := len(part)
	for _, sw := range []string{" limit ", " offset ", ";"} {
		if i := strings.Index(part, sw); i >= 0 && i < end {
			end = i
		}
	}
	cols := strings.Split(part[:end], ",")
	out := []string{}
	for _, c := range cols {
		c = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(c, " desc"), " asc"))
		c = strings.Trim(c, "() ")
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

func extractColumnRefs(s string) []string {
	re := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_\.]*)(\s*=|\s+in\s*\(|\s+between\s+|\s+>|\s+<|\s+like\s+|\s+ilike\s+|\s+is\s+)`)
	matches := re.FindAllStringSubmatch(s, -1)
	out := []string{}
	for _, m := range matches {
		col := strings.TrimSpace(m[1])
		if !isKeyword(col) {
			out = append(out, col)
		}
	}
	return out
}

func isKeyword(s string) bool {
	switch s {
	case "and", "or", "not", "case", "when", "then", "else", "end":
		return true
	}
	return false
}

func unique(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

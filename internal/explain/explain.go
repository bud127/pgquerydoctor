package explain

import (
	"regexp"
	"strconv"
	"strings"
)

type PlanInfo struct {
	Raw               string
	HasSeqScan        bool
	HasIndexScan      bool
	HasBitmapHeapScan bool
	HasNestedLoop     bool
	HasHashJoin       bool
	HasSort           bool
	HasExternalMerge  bool
	HasTempIO         bool
	HasBuffers        bool
	PlanningTimeMS    float64
	ExecutionTimeMS   float64
	SeqScanTables     []string
	HighRowNodes      []string
}

func Parse(raw string) PlanInfo {
	lower := strings.ToLower(raw)
	pi := PlanInfo{Raw: raw}
	pi.HasSeqScan = strings.Contains(lower, "seq scan")
	pi.HasIndexScan = strings.Contains(lower, "index scan") || strings.Contains(lower, "index only scan")
	pi.HasBitmapHeapScan = strings.Contains(lower, "bitmap heap scan")
	pi.HasNestedLoop = strings.Contains(lower, "nested loop")
	pi.HasHashJoin = strings.Contains(lower, "hash join")
	pi.HasSort = strings.Contains(lower, "sort")
	pi.HasExternalMerge = strings.Contains(lower, "external merge") || strings.Contains(lower, "disk:")
	pi.HasTempIO = strings.Contains(lower, "temp read") || strings.Contains(lower, "temp written") || strings.Contains(lower, "temp")
	pi.HasBuffers = strings.Contains(lower, "buffers:")
	pi.PlanningTimeMS = readMS(lower, `planning time:\s*([0-9\.]+)\s*ms`)
	pi.ExecutionTimeMS = readMS(lower, `execution time:\s*([0-9\.]+)\s*ms`)
	pi.SeqScanTables = extractSeqScanTables(lower)
	pi.HighRowNodes = extractHighRows(raw)
	return pi
}

func readMS(s, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	return v
}

func extractSeqScanTables(s string) []string {
	re := regexp.MustCompile(`seq scan on\s+([a-zA-Z0-9_\.]+)`)
	matches := re.FindAllStringSubmatch(s, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

func extractHighRows(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := []string{}
	re := regexp.MustCompile(`rows=([0-9]+)`)
	for _, line := range lines {
		m := re.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		rows, _ := strconv.Atoi(m[1])
		if rows >= 100000 {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

package explain

import "testing"

func TestParseExplain(t *testing.T) {
	raw := `Seq Scan on users (cost=0.00..100.00 rows=200000 width=10) (actual time=1.00..2.00 rows=200000 loops=1)
Sort Method: external merge Disk: 1024kB
Planning Time: 1.234 ms
Execution Time: 55.678 ms`
	pi := Parse(raw)
	if !pi.HasSeqScan || !pi.HasExternalMerge || pi.ExecutionTimeMS != 55.678 || len(pi.SeqScanTables) != 1 {
		t.Fatalf("unexpected explain parse: %+v", pi)
	}
}

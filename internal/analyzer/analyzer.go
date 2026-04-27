package analyzer

import (
	"github.com/budisuryadi/pgquerydoctor/internal/explain"
	"github.com/budisuryadi/pgquerydoctor/internal/parser"
	"github.com/budisuryadi/pgquerydoctor/internal/rules"
)

type Severity = rules.Severity
type Finding = rules.Finding

type Result struct {
	Query    parser.QueryInfo
	Plan     explain.PlanInfo
	Findings []Finding
}

func Analyze(sql, plan string) Result {
	qi := parser.Parse(sql)
	pi := explain.Parse(plan)
	ctx := rules.Context{Query: qi, Plan: pi}
	findings := rules.Run(ctx)
	return Result{Query: qi, Plan: pi, Findings: findings}
}

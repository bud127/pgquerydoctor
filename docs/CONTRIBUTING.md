# Contributing

Thanks for helping improve PgQueryDoctor.

## Development

```bash
go test ./...
go run ./cmd/pgquerydoctor analyze --query examples/query.sql --explain examples/explain.txt
```

## Rule contribution guide

A good rule should include:

1. Clear trigger condition
2. Evidence from SQL or EXPLAIN
3. Practical recommendation
4. Risk or tradeoff
5. Test case

Please keep recommendations conservative. PgQueryDoctor should suggest, not blindly claim.

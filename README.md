# PgQueryDoctor

PgQueryDoctor is an open source PostgreSQL query analyzer for backend engineers.
It explains slow queries, detects common performance bottlenecks, and suggests practical indexes or query rewrites.

> Status: early open source MVP. Rule-based first, not AI-dependent.

## What problem this solves

Backend engineers often receive a slow SQL query and a long `EXPLAIN ANALYZE` output, but it is hard to quickly answer:

- Why is this query slow?
- Is PostgreSQL using the right index?
- Is the query doing a large sequential scan?
- Is sorting spilling to disk?
- Is `OFFSET` pagination becoming expensive?
- What index or rewrite should be tested next?

PgQueryDoctor gives a simple Markdown report with evidence, likely root cause, recommended fix, tradeoff, and suggested indexes.

## Who this is for

- Backend engineers working with PostgreSQL
- SRE / DevOps engineers investigating slow queries
- Tech leads reviewing database performance
- Open source maintainers who want a simple CLI analyzer

## Features

- Paste or pass SQL query from file
- Paste or pass `EXPLAIN ANALYZE` output from file
- Detect SQL shape:
  - `SELECT`
  - `JOIN`
  - `WHERE`
  - `GROUP BY`
  - `ORDER BY`
  - `LIMIT`
  - `OFFSET`
  - `CTE`
  - subquery
  - `DISTINCT ON`
  - JSONB operators
- Analyze EXPLAIN text:
  - Seq Scan
  - Index Scan
  - Bitmap Heap Scan
  - Nested Loop
  - Hash Join
  - Sort
  - External Merge Sort
  - Temp Read/Write
  - Buffers
  - Planning Time
  - Execution Time
- Generate Markdown report
- CLI-first
- Docker support
- GitHub Actions CI
- Unit tests

## Installation

```bash
git clone https://github.com/budisuryadi/pgquerydoctor.git
cd pgquerydoctor
go build -o pgquerydoctor ./cmd/pgquerydoctor
./pgquerydoctor version
```

## CLI Usage

```bash
./pgquerydoctor analyze --query examples/query.sql --explain examples/explain.txt
```

Generate Markdown report:

```bash
./pgquerydoctor report --query examples/query.sql --explain examples/explain.txt --output report.md
```

Lint SQL only:

```bash
./pgquerydoctor lint --query examples/query.sql
```

Suggest indexes only:

```bash
./pgquerydoctor suggest-index --query examples/query.sql
```

## Docker Usage

```bash
docker build -t pgquerydoctor .
docker run --rm -v "$PWD/examples:/workspace" pgquerydoctor analyze --query /workspace/query.sql --explain /workspace/explain.txt
```

## Example Inputs

The default example uses a generic slow pagination query:

```sql
SELECT id, full_name, email, created_at
FROM users
WHERE status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;
```

Additional examples are available in `examples/`:

- `slow-pagination.sql` — deep `OFFSET` pagination
- `jsonb-filter.sql` — JSONB filter using `@>`
- `distinct-on.sql` — latest row per customer with `DISTINCT ON`
- `heavy-groupby.sql` — aggregation with `GROUP BY` and `HAVING`
- `nested-loop.sql` — join plan with high-loop nested loop

## Example Report Output

```markdown
# PgQueryDoctor Report

## Summary

- Tables detected: `users`
- Planning time: `1.220 ms`
- Execution time: `421.004 ms`
- Findings: `4`

## Bottlenecks and Recommendations

### Deep OFFSET pagination detected [WARN]

Evidence: SQL contains OFFSET 10000.

Recommended Fix: Replace deep OFFSET pagination with keyset pagination using the last seen cursor values.

Suggested Index:

CREATE INDEX CONCURRENTLY idx_users_status_created_at_id
ON users (status, created_at DESC, id DESC);
```

## Supported PostgreSQL Patterns

Current MVP supports common patterns:

- Large sequential scan
- External merge sort / disk sort
- Nested Loop with high row count
- JSONB filter indexing strategy
- `ORDER BY + LIMIT`
- `OFFSET` pagination
- `DISTINCT ON`
- `GROUP BY + HAVING MAX(CASE WHEN...)`
- Potential composite index for `WHERE + ORDER BY`
- CTE detection

## Design Philosophy

PgQueryDoctor is conservative by default.
It does not say an index is always correct. It suggests what to test next with:

```sql
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
```

Recommended indexes should be validated on production-like data before deployment.

## Project Structure

```text
cmd/pgquerydoctor          CLI entrypoint
internal/analyzer        Analysis orchestration
internal/parser          Lightweight SQL shape parser
internal/explain         EXPLAIN text parser
internal/rules           Rule engine
internal/report          Markdown report generator
examples                 Sample SQL and EXPLAIN files
docs                     Installation, roadmap, contributing guide
tests                    Integration-style tests
```

## Before / After Analysis Example

Before:

```sql
SELECT id, full_name, email, created_at
FROM users
WHERE status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;
```

Problem:

- Deep `OFFSET` gets slower as page number grows.
- PostgreSQL still needs to scan, sort, or walk skipped rows.
- The plan may show `Seq Scan`, expensive `Sort`, temp read/write, or disk spill.

After:

```sql
SELECT id, full_name, email, created_at
FROM users
WHERE status = 'ACTIVE'
  AND (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

Suggested index to test:

```sql
CREATE INDEX CONCURRENTLY idx_users_status_created_at_id
ON users (status, created_at DESC, id DESC);
```

This is keyset pagination. It is usually faster for infinite scroll and large datasets.

## Roadmap

- JSON EXPLAIN parser
- Schema-aware index recommendations
- Table statistics support
- Configurable rule thresholds
- REST API
- Web UI
- GitHub PR bot
- PostgreSQL version-specific rules

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md).

## License

MIT License is recommended for maximum adoption and easier open source contribution.

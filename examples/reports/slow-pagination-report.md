# PgQueryDoctor Report

## Summary

- Tables detected: `users`
- Planning time: `1.220 ms`
- Execution time: `421.004 ms`
- Findings: `4`

## Bottlenecks and Recommendations

### Deep OFFSET pagination detected [WARN]

Evidence: SQL contains `OFFSET 10000`.

Likely Root Cause: PostgreSQL still has to scan, sort, or walk skipped rows before returning the requested page.

Recommended Fix: Replace deep OFFSET with keyset pagination using the last seen `created_at` and `id` values.

Suggested Index:

```sql
CREATE INDEX CONCURRENTLY idx_users_status_created_at_id
ON users (status, created_at DESC, id DESC);
```

Optimized SQL Example:

```sql
SELECT id, full_name, email, created_at
FROM users
WHERE status = 'ACTIVE'
  AND (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

Risk / Tradeoff: Keyset pagination is excellent for next/previous navigation but not ideal for jumping directly to page 500.

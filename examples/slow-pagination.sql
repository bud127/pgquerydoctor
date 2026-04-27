SELECT id, full_name, email, created_at
FROM users
WHERE status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;

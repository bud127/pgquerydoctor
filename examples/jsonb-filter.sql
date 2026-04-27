SELECT id, payload, created_at
FROM orders
WHERE payload @> '{"status":"PAID","channel":"MOBILE"}'
ORDER BY created_at DESC
LIMIT 50;

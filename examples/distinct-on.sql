SELECT DISTINCT ON (customer_id)
       customer_id, id, total_amount, created_at
FROM orders
WHERE status = 'PAID'
ORDER BY customer_id, created_at DESC;

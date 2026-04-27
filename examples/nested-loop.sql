SELECT o.id, o.customer_id, p.payment_method, p.status
FROM orders o
JOIN payments p ON p.order_id = o.id
WHERE o.created_at >= NOW() - INTERVAL '30 days';

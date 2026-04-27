SELECT customer_id,
       COUNT(*) AS total_orders,
       SUM(total_amount) AS revenue
FROM orders
WHERE created_at >= NOW() - INTERVAL '90 days'
GROUP BY customer_id
HAVING COUNT(*) > 10
ORDER BY revenue DESC;

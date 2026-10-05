-- One approved PR may produce only one PO, and one PO may consume budget once.
-- Run after resolving any legacy duplicates reported by the SELECT statements
-- below; the unique indexes intentionally reject ambiguous historical data.

SELECT pr_id, COUNT(*) AS duplicate_count
FROM purchase_orders
WHERE pr_id IS NOT NULL
GROUP BY pr_id
HAVING COUNT(*) > 1;

SELECT ref_type, ref_id, COUNT(*) AS duplicate_count
FROM budget_usages
GROUP BY ref_type, ref_id
HAVING COUNT(*) > 1;

ALTER TABLE purchase_orders
    ADD UNIQUE KEY uq_purchase_orders_pr_id (pr_id);

ALTER TABLE budget_usages
    ADD UNIQUE KEY uq_budget_usages_reference (ref_type, ref_id);

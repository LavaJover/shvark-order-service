-- +migrate Down
-- +migrate DisableTransaction

-- Удаление индексов. DROP INDEX CONCURRENTLY не существует, но это быстрая операция.
-- Можно выполнить в транзакции.
DROP INDEX IF EXISTS idx_order_pending;
DROP INDEX IF EXISTS idx_order_composite_stats;
DROP INDEX IF EXISTS idx_bank_detail_composite;
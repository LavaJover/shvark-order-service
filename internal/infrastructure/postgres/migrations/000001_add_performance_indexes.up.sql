-- +migrate Up
-- +migrate DisableTransaction

-- Для CREATE INDEX CONCURRENTLY отключаем транзакцию
-- +migrate StatementBegin
CREATE INDEX IF NOT EXISTS idx_bank_detail_composite ON bank_detail_models (enabled, currency, payment_system, min_amount, max_amount) WHERE deleted_at IS NULL;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE INDEX IF NOT EXISTS idx_order_composite_stats ON order_models (bank_details_id, status, created_at, amount_fiat);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE INDEX IF NOT EXISTS idx_order_pending ON order_models (bank_details_id, amount_fiat) WHERE status = 'PENDING';
-- +migrate StatementEnd
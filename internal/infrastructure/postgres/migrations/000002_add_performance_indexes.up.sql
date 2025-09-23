-- +migrate Up

-- Критически важные индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_bank_detail_static_filters 
ON bank_detail_models (enabled, currency, payment_system, min_amount, max_amount) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_order_bank_status_created 
ON order_models (bank_details_id, status, created_at);

CREATE INDEX IF NOT EXISTS idx_order_status_amount 
ON order_models (bank_details_id, status, amount_fiat);

CREATE INDEX IF NOT EXISTS idx_order_pending_amount 
ON order_models (bank_details_id, amount_fiat) 
WHERE status = 'PENDING';

-- Индекс для быстрого поиска по датам
CREATE INDEX IF NOT EXISTS idx_order_created_at 
ON order_models (created_at) 
WHERE status IN ('PENDING', 'COMPLETED');
-- Основной индекс для поиска pending orders по device_id
CREATE INDEX idx_orders_pending_device 
ON order_models (status, expires_at) 
WHERE status = 'PENDING' AND expires_at > NOW();

-- Индекс для поиска по device_id через join с bank_details
CREATE INDEX idx_orders_bank_details_id 
ON order_models (bank_details_id) 
WHERE status = 'PENDING';

-- Индекс для поиска по сумме (с фильтром по статусу)
CREATE INDEX idx_orders_amount_pending 
ON order_models (amount_fiat, status) 
WHERE status = 'PENDING';

-- Индекс для временных диапазонов
CREATE INDEX idx_orders_created_at 
ON order_models (created_at DESC);

-- Композитный индекс для основных запросов
CREATE INDEX idx_orders_status_expires_created 
ON order_models (status, expires_at, created_at);

-- Основной индекс для поиска по device_id
CREATE INDEX idx_bank_details_device_id 
ON bank_detail_models (device_id) 
WHERE enabled = true AND deleted_at IS NULL;

-- Индекс для поиска активных реквизитов
CREATE INDEX idx_bank_details_active 
ON bank_detail_models (trader_id, enabled, currency) 
WHERE enabled = true AND deleted_at IS NULL;

-- Индекс для поиска по платежным системам
CREATE INDEX idx_bank_details_payment_system 
ON bank_detail_models (payment_system, enabled) 
WHERE enabled = true;

-- Основной уникальный индекс для идемпотентности
CREATE UNIQUE INDEX CONCURRENTLY idx_payment_log_order_hash 
ON payment_processing_logs (order_id, payment_hash);

-- Индекс для быстрого поиска по хешу платежа
CREATE INDEX CONCURRENTLY idx_payment_log_hash 
ON payment_processing_logs (payment_hash);

-- Индекс для временных запросов и очистки
CREATE INDEX CONCURRENTLY idx_payment_log_processed_at 
ON payment_processing_logs (processed_at DESC);

-- Индекс для отчетности и мониторинга
CREATE INDEX idx_payment_log_success_processed 
ON payment_processing_logs (success, processed_at);

CREATE INDEX idx_order_tx_order_id 
ON order_transaction_states (order_id);

CREATE INDEX idx_order_tx_operation_created 
ON order_transaction_states (operation, created_at);
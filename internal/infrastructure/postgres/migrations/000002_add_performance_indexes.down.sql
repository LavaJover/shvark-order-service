-- +migrate Down

DROP INDEX IF EXISTS idx_bank_detail_static_filters;
DROP INDEX IF EXISTS idx_order_bank_status_created;
DROP INDEX IF EXISTS idx_order_status_amount;
DROP INDEX IF EXISTS idx_order_pending_amount;
DROP INDEX IF EXISTS idx_order_created_at;
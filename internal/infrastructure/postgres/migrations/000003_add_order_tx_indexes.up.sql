-- +migrate up

-- Индексы для работы с состоянием транзакций сделок
CREATE INDEX idx_order_transaction_states_order_id ON order_transaction_states(order_id);
CREATE INDEX idx_order_transaction_states_operation ON order_transaction_states(operation);
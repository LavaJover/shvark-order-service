-- +migrate up

-- Индексы для быстрого поиска проблемных ордеров
CREATE INDEX idx_order_processing_status ON order_models(status, released_at, published_at);
CREATE INDEX idx_order_release_attempts ON order_models(release_attempts) WHERE release_attempts > 0;
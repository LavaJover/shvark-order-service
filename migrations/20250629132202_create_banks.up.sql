CREATE TABLE banks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT UNIQUE NOT NULL,     -- "sberbank"
    name TEXT NOT NULL            -- "Сбербанк"
);

CREATE TABLE payment_details (
    id UUID PRIMARY KEY,
    bank_id UUID NOT NULL REFERENCES banks(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('SBP', 'C2C')),

    phone_number TEXT,
    card_number TEXT,

    CHECK (
        phone_number IS NOT NULL OR
        card_number IS NOT NULL
    ),

    min_amount NUMERIC(18, 2) NOT NULL,
    max_amount NUMERIC(18, 2) NOT NULL,

    min_mounthly_amount NUMERIC(18, 2),
    max_mounthly_amount NUMERIC(18, 2),

    max_active_orders INTEGER NOT NULL,

    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_payment_details_bank_id ON payment_details(bank_id);
CREATE INDEX idx_payment_details_type ON payment_details(type);
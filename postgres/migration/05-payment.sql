CREATE SCHEMA payment;

CREATE TABLE payment.payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES sales.orders(id),

    amount          NUMERIC(12, 2) NOT NULL,
    currency        CHAR(3) NOT NULL,

    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT payments_amount_positive
        CHECK (amount > 0),

    CONSTRAINT payments_currency_valid
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT payments_status_valid
        CHECK (
            status IN (
                'PENDING',
                'COMPLETED',
                'FAILED',
                'CANCELLED',
                'REFUNDED'
            )
        )

);

CREATE INDEX idx_payments_status
    ON payment.payments (status);

CREATE INDEX idx_payments_created_at
    ON payment.payments (created_at);
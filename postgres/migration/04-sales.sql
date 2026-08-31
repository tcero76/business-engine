CREATE SCHEMA sales;

CREATE TABLE sales.orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id         UUID NOT NULL REFERENCES demo.users(id),

    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    total_amount    NUMERIC(12, 2) NOT NULL,
    currency        CHAR(3) NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT orders_total_positive
        CHECK (total_amount >= 0),

    CONSTRAINT orders_currency_valid
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT orders_status_valid
        CHECK (
            status IN (
                'PENDING',
                'CANCELLED',
                'PAID',
                'COMPLETED'
            )
        )
);

CREATE INDEX idx_orders_user_id
    ON sales.orders (user_id);

CREATE INDEX idx_orders_status
    ON sales.orders (status);

CREATE INDEX idx_orders_created_at
    ON sales.orders (created_at);

CREATE TABLE sales.order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id        UUID NOT NULL REFERENCES sales.orders(id),

    appointment_id UUID REFERENCES scheduling.appointments(id),
    product_id     UUID REFERENCES inventory.products(id),

    quantity        INTEGER NOT NULL DEFAULT 1,

    unit_price      NUMERIC(12, 2) NOT NULL,
    total_amount    NUMERIC(12, 2) NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT order_items_quantity_positive
        CHECK (quantity > 0),

    CONSTRAINT order_items_unit_price_positive
        CHECK (unit_price > 0),

    CONSTRAINT order_items_total_positive
        CHECK (total_amount > 0),

    CONSTRAINT order_items_single_source
        CHECK (
            (appointment_id IS NOT NULL AND product_id IS NULL)
            OR
            (appointment_id IS NULL AND product_id IS NOT NULL)
        )
);

CREATE INDEX idx_order_items_order_id
    ON sales.order_items (order_id);

CREATE INDEX idx_order_items_appointment_id
    ON sales.order_items (appointment_id);

CREATE INDEX idx_order_items_product_id
    ON sales.order_items (product_id);

CREATE TABLE sales.return_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id        UUID NOT NULL REFERENCES sales.orders(id),
    user_id         UUID NOT NULL REFERENCES demo.users(id),

    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    total_amount    NUMERIC(12, 2) NOT NULL,
    currency        CHAR(3) NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT return_orders_total_positive
        CHECK (total_amount > 0),

    CONSTRAINT return_orders_currency_valid
        CHECK (currency ~ '^[A-Z]{3}$'),

    CONSTRAINT return_orders_status_valid
        CHECK (
            status IN (
                'PENDING',
                'CANCELLED',
                'COMPLETED'
            )
        )
);

CREATE INDEX idx_return_orders_order_id
    ON sales.return_orders (order_id);

CREATE INDEX idx_return_orders_user_id
    ON sales.return_orders (user_id);

CREATE INDEX idx_return_orders_status
    ON sales.return_orders (status);

CREATE TABLE sales.return_order_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    return_order_id     UUID NOT NULL
        REFERENCES sales.return_orders(id),

    order_item_id       UUID NOT NULL
        REFERENCES sales.order_items(id),

    quantity            INTEGER NOT NULL,

    unit_price          NUMERIC(12, 2) NOT NULL,
    total_amount        NUMERIC(12, 2) NOT NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT return_order_items_quantity_positive
        CHECK (quantity > 0),

    CONSTRAINT return_order_items_unit_price_positive
        CHECK (unit_price > 0),

    CONSTRAINT return_order_items_total_positive
        CHECK (total_amount > 0)
);

CREATE INDEX idx_return_order_items_return_order_id
    ON sales.return_order_items (return_order_id);

CREATE INDEX idx_return_order_items_order_item_id
    ON sales.return_order_items (order_item_id);
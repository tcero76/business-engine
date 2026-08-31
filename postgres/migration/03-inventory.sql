CREATE SCHEMA inventory;

CREATE TABLE inventory.products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    sku         VARCHAR(100) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    price NUMERIC(12, 2) NOT NULL,

    active      BOOLEAN NOT NULL DEFAULT TRUE,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT products_price_positive
        CHECK (price > 0),
    CONSTRAINT products_sku_unique
        UNIQUE (sku)
);


CREATE TABLE inventory.stock (
    product_id  UUID PRIMARY KEY REFERENCES inventory.products(id),

    quantity    INTEGER NOT NULL DEFAULT 0,

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT stock_quantity_non_negative
        CHECK (quantity >= 0)
);


CREATE TABLE inventory.inventory_movements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id  UUID NOT NULL REFERENCES inventory.products(id),
    user_id     UUID NOT NULL REFERENCES demo.users(id),

    movement_type VARCHAR(20) NOT NULL,

    quantity    INTEGER NOT NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT inventory_movements_type_valid
        CHECK (
            movement_type IN (
                'PURCHASE',
                'SALE',
                'ADJUSTMENT'
            )
        ),

    CONSTRAINT inventory_movements_quantity_positive
        CHECK (quantity > 0)
);

CREATE INDEX idx_inventory_movements_product
    ON inventory.inventory_movements (product_id, created_at);

CREATE INDEX idx_inventory_movements_user
    ON inventory.inventory_movements (user_id);
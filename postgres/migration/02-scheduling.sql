CREATE SCHEMA scheduling;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE scheduling.resources (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    price       NUMERIC(12, 2) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    CONSTRAINT resources_price_positive
        CHECK (price > 0)
);

CREATE TABLE scheduling.availability (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES scheduling.resources(id),

    available_period TSTZRANGE NOT NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT availability_period_valid
        CHECK (
            NOT isempty(available_period)
            AND NOT lower_inf(available_period)
            AND NOT upper_inf(available_period)
        )
);

ALTER TABLE scheduling.availability
ADD CONSTRAINT availability_no_overlap
EXCLUDE USING GIST (
    resource_id WITH =,
    available_period WITH &&
);


CREATE TABLE scheduling.appointments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    resource_id UUID NOT NULL REFERENCES scheduling.resources(id),
    user_id     UUID NOT NULL REFERENCES demo.users(id),

    appointment_period TSTZRANGE NOT NULL,

    status      VARCHAR(20) NOT NULL DEFAULT 'CONFIRMED',

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT appointments_period_valid
        CHECK (
            NOT isempty(appointment_period)
            AND NOT lower_inf(appointment_period)
            AND NOT upper_inf(appointment_period)
        ),

    CONSTRAINT appointments_status_valid
        CHECK (
            status IN (
                'CONFIRMED',
                'CANCELLED',
                'COMPLETED'
            )
        )
);

CREATE INDEX idx_appointments_resource
    ON scheduling.appointments (resource_id);

CREATE INDEX idx_appointments_user
    ON scheduling.appointments (user_id);

CREATE INDEX idx_appointments_period
    ON scheduling.appointments
    USING GIST (appointment_period);


ALTER TABLE scheduling.appointments
ADD CONSTRAINT appointments_no_overlap
EXCLUDE USING GIST (
    resource_id WITH =,
    appointment_period WITH &&
)
WHERE (status = 'CONFIRMED');
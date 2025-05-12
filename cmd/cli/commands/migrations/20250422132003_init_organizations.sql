-- +goose Up
-- +goose StatementBegin
-- Create the necessary extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the billing_plan_type enum if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'billing_plan_type') THEN
        CREATE TYPE billing_plan_type AS ENUM (
            '10_DEVICES_99_MONTH',
            '50_DEVICES_399_MONTH',
            '100_DEVICES_799_MONTH'
        );
    END IF;
END$$;

-- Create the organizations table if it doesn't exist
CREATE TABLE IF NOT EXISTS organizations (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    primary_administrator_email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    billing_plan_type billing_plan_type NOT NULL,
    payment_details JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organizations;
DROP TYPE IF EXISTS billing_plan_type;
-- We don't drop the uuid-ossp extension as it might be used by other parts of the database
-- +goose StatementEnd
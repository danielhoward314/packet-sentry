-- +goose Up
-- +goose StatementBegin
-- Create the necessary extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'key_hash_type') THEN
        CREATE TYPE key_hash_type AS ENUM (
            'SHA256'
        );
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS install_keys (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    key_hash TEXT NOT NULL,
    key_hash_type key_hash_type NOT NULL,
    administrator_id UUID NOT NULL,
    CONSTRAINT fk_administrator
        FOREIGN KEY(administrator_id) 
        REFERENCES administrators(id)
        ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS install_keys;
DROP TYPE IF EXISTS key_hash_type;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'password_hash_type') THEN
        CREATE TYPE password_hash_type AS ENUM (
            'BCRYPT'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'authorization_role') THEN
        CREATE TYPE authorization_role AS ENUM (
            'PRIMARY_ADMIN',
            'SECONDARY_ADMIN'
        );
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS administrators (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    password_hash_type password_hash_type NOT NULL,
    authorization_role authorization_role NOT NULL,
    password_hash TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT false,
    organization_id UUID NOT NULL,
    CONSTRAINT fk_organization
        FOREIGN KEY(organization_id) 
        REFERENCES organizations(id)
        ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS administrators;
DROP TYPE IF EXISTS authorization_role;
DROP TYPE IF EXISTS password_hash_type;
-- +goose StatementEnd
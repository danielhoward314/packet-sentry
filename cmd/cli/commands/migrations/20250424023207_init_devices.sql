-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS devices (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    os_unique_identifier TEXT NOT NULL UNIQUE,
    client_cert_pem TEXT NOT NULL,
    client_cert_fingerprint TEXT NOT NULL,
    organization_id UUID NOT NULL,
    CONSTRAINT fk_organization
        FOREIGN KEY(organization_id) 
        REFERENCES organizations(id)
        ON DELETE CASCADE,
    interface_bpf_associations JSONB DEFAULT '{}'::jsonb,
    previous_associations JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_os_unique_identifier ON devices(os_unique_identifier);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_devices_os_unique_identifier;
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd

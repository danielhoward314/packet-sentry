package queries

const DevicesInsert = `
INSERT INTO devices (os_unique_identifier, client_cert_pem, client_cert_fingerprint,
					 organization_id)
VALUES ($1, $2, $3, $4)
RETURNING id
`

const DevicesSelectByOSUniqueIdentifier = `
SELECT id, os_unique_identifier, client_cert_pem, client_cert_fingerprint,
       organization_id, interface_bpf_associations, previous_associations
FROM devices
WHERE os_unique_identifier = $1
`

const DevicesUpdate = `
UPDATE devices
SET client_cert_pem = $1,
	client_cert_fingerprint = $2,
	interface_bpf_associations = $3,
	previous_associations = $4
WHERE id = $5
RETURNING id
`

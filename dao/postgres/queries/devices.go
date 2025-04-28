package queries

const DevicesInsert = `
INSERT INTO devices (os_unique_identifier, client_cert_pem, client_cert_fingerprint,
					 organization_id)
VALUES ($1, $2, $3, $4)
RETURNING id
`

const DevicesSelectById = `
SELECT id, os_unique_identifier, client_cert_pem, client_cert_fingerprint, organization_id,
       pcap_version, interfaces, interface_bpf_associations, previous_associations
FROM devices
WHERE id = $1
`

const DevicesSelectByOSUniqueIdentifier = `
SELECT id, os_unique_identifier, client_cert_pem, client_cert_fingerprint, organization_id,
       pcap_version, interfaces, interface_bpf_associations, previous_associations
FROM devices
WHERE os_unique_identifier = $1
`

const DevicesSelectByOrganizationID = `
SELECT id, os_unique_identifier, client_cert_pem, client_cert_fingerprint, organization_id,
       pcap_version, interfaces, interface_bpf_associations, previous_associations
FROM devices
WHERE organization_id = $1
`

const DevicesUpdate = `
UPDATE devices
SET client_cert_pem = $1,
	client_cert_fingerprint = $2,
	pcap_version = $3,
	interfaces = $4,
	interface_bpf_associations = $5,
	previous_associations = $6
WHERE id = $7
RETURNING id
`

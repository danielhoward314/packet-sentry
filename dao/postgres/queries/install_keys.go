package queries

const InstallKeysInsert = `
INSERT INTO install_keys (key_hash, key_hash_type, administrator_id, organization_id)
VALUES ($1, $2, $3, $4)
RETURNING id
`

const InstallKeysSelect = `
SELECT administrator_id, organization_id
FROM install_keys
WHERE key_hash = $1`

const InstallKeysDelete = `DELETE FROM install_keys where key_hash = $1`

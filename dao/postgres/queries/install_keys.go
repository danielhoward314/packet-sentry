package queries

const InstallKeysInsert = `INSERT INTO install_keys (key_hash, key_hash_type, administrator_id)
VALUES ($1, $2, $3)
RETURNING id
`

const InstallKeysSelect = `SELECT id FROM install_keys where key_hash = $1`

const InstallKeysDelete = `DELETE FROM install_keys where key_hash = $1`

package queries

const OrganizationsInsert = `INSERT INTO organizations (primary_administrator_email, name, billing_plan_type)
VALUES ($1, $2, $3)
RETURNING id
`

const OrganizationsSelect = `SELECT
	id, primary_administrator_email, name, billing_plan_type
FROM organizations
WHERE id = $1`

const OrganizationsUpdate = `UPDATE organizations
SET
    name = $1,
    billing_plan_type = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $3
`

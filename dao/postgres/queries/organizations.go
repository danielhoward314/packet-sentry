package queries

const OrganizationsInsert = `INSERT INTO organizations (primary_administrator_email, name, billing_plan_type)
VALUES ($1, $2, $3)
RETURNING id
`

const OrganizationsSelect = `SELECT id, primary_administrator_email, name, billing_plan_type FROM organizations where id = $1`

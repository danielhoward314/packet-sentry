package postgres

import (
	"database/sql"

	"github.com/danielhoward314/packet-sentry/dao"
)

// NewDatastore returns a postgres implementation for the primary datastore
func NewDatastore(db *sql.DB) *dao.Datastore {
	return &dao.Datastore{
		Administrators: NewAdministrators(db),
		Organizations:  NewOrganizations(db),
	}
}

package postgres

import (
	"database/sql"

	"github.com/danielhoward314/packet-sentry/dao"
)

// NewDatastore returns a postgres implementation for the primary datastore
func NewDatastore(db *sql.DB, installKeySecret string) *dao.Datastore {
	return &dao.Datastore{
		Administrators: NewAdministrators(db),
		InstallKeys:    NewInstallKeys(db, installKeySecret),
		Organizations:  NewOrganizations(db),
	}
}

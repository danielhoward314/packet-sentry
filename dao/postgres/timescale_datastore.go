package postgres

import (
	"database/sql"

	"github.com/danielhoward314/packet-sentry/dao"
)

// NewTimescaleDatastore returns a postgres implementation for the timescale datastore
func NewTimescaleDatastore(db *sql.DB) *dao.TimescaleDatastore {
	return &dao.TimescaleDatastore{
		Events: NewEvents(db),
	}
}

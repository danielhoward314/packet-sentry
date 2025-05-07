package postgres

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
	"github.com/danielhoward314/packet-sentry/hashes"
)

type administrators struct {
	db *sql.DB
}

// AuthorizationRole is a type alias representing the postgres ENUM for the authorization_role column
type AuthorizationRole string

const (
	// PrimaryAdmin is a string for the primary admin authorization_role ENUM
	PrimaryAdmin = "PRIMARY_ADMIN"

	// SecondaryAdmin is a string for the secondary admin authorization_role ENUM
	SecondaryAdmin = "SECONDARY_ADMIN"
)

// PasswordHashType is a type alias representing the postgres ENUM for the password_hash_type column
type PasswordHashType string

const (
	// PasswordHashTypeBCrypt is a string for the bcrypt password_hash_type ENUM
	PasswordHashTypeBCrypt = "BCRYPT"
)

// NewAdministrators returns an instance implementing the Administrators interface
func NewAdministrators(db *sql.DB) dao.Administrators {
	return &administrators{db: db}
}

func (a *administrators) Create(administrator *dao.Administrator, primaryAdministratorCleartextPassword string) (string, error) {
	if administrator == nil {
		return "", errors.New("invalid administrator")
	}
	if administrator.Email == "" {
		return "", errors.New("invalid administrator email")
	}
	if administrator.DisplayName == "" {
		return "", errors.New("invalid administrator display name")
	}
	if administrator.OrganizationID == "" {
		return "", errors.New("invalid administrator organization id")
	}
	if primaryAdministratorCleartextPassword == "" {
		return "", errors.New("invalid administrator password cleartext")
	}
	passwordHash, err := hashes.HashCleartextWithBCrypt(primaryAdministratorCleartextPassword)
	if err != nil {
		return "", err
	}
	var id string
	err = a.db.QueryRow(
		queries.AdministratorsInsert,
		administrator.Email,
		administrator.DisplayName,
		administrator.OrganizationID,
		PasswordHashTypeBCrypt,
		passwordHash,
		administrator.AuthorizationRole,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *administrators) Read(id string) (*dao.Administrator, error) {
	administrator := &dao.Administrator{}
	err := a.db.QueryRow(queries.AdministratorsSelect, id).Scan(
		&administrator.ID,
		&administrator.Email,
		&administrator.DisplayName,
		&administrator.PasswordHashType,
		&administrator.PasswordHash,
		&administrator.OrganizationID,
		&administrator.Verified,
		&administrator.AuthorizationRole,
	)
	if err != nil {
		return nil, err
	}
	return administrator, nil
}

func (a *administrators) ReadByEmail(email string) (*dao.Administrator, error) {
	administrator := &dao.Administrator{}
	err := a.db.QueryRow(queries.AdministratorsSelectByEmail, email).Scan(
		&administrator.ID,
		&administrator.Email,
		&administrator.DisplayName,
		&administrator.PasswordHashType,
		&administrator.PasswordHash,
		&administrator.OrganizationID,
		&administrator.Verified,
		&administrator.AuthorizationRole,
	)
	if err != nil {
		return nil, err
	}
	return administrator, nil
}

func (a *administrators) List(organizationID string) ([]*dao.Administrator, error) {
	if organizationID == "" {
		return nil, errors.New("empty organization id")
	}

	var administrators []*dao.Administrator
	rows, rowsErr := a.db.Query(queries.AdministratorsSelectByOrganizationID, organizationID)
	if rowsErr != nil {
		return nil, rowsErr
	}

	for rows.Next() {
		var administrator dao.Administrator
		rowErr := rows.Scan(
			&administrator.ID,
			&administrator.Email,
			&administrator.DisplayName,
			&administrator.PasswordHashType,
			&administrator.PasswordHash,
			&administrator.OrganizationID,
			&administrator.Verified,
			&administrator.AuthorizationRole,
		)
		if rowErr != nil {
			return nil, rowErr
		}

		administrators = append(administrators, &administrator)
	}

	return administrators, nil
}

func (a *administrators) Update(administrator *dao.Administrator) error {
	if administrator == nil {
		return errors.New("invalid administrator")
	}
	_, err := a.db.Exec(
		queries.AdministratorsUpdate,
		administrator.Email,
		administrator.DisplayName,
		administrator.PasswordHashType,
		administrator.PasswordHash,
		administrator.OrganizationID,
		administrator.Verified,
		administrator.AuthorizationRole,
		administrator.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (a *administrators) Delete(id string) (int64, error) {
	result, err := a.db.Exec(queries.AdministratorsDelete, id)
	if err != nil {
		return 0, err
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsDeleted, nil
}

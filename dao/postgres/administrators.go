package postgres

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
	"github.com/danielhoward314/packet-sentry/passwords"
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
	// BCryptHashType is a string for the bcrypt password_hash_type ENUM
	BCryptHashType = "BCRYPT"
)

// NewAdministrators returns an instance implementing the Administrators interface
func NewAdministrators(db *sql.DB) dao.Administrators {
	return &administrators{db: db}
}

func (o *administrators) Create(administrator *dao.Administrator, primaryAdministratorCleartextPassword string) (string, error) {
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
	passwordHash, err := passwords.HashPasswordWithBCrypt(primaryAdministratorCleartextPassword)
	if err != nil {
		return "", err
	}
	var id string
	err = o.db.QueryRow(
		queries.AdministratorsInsert,
		administrator.Email,
		administrator.DisplayName,
		administrator.OrganizationID,
		BCryptHashType,
		passwordHash,
		administrator.AuthorizationRole,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (o *administrators) Read(id string) (*dao.Administrator, error) {
	administrator := &dao.Administrator{}
	err := o.db.QueryRow(queries.AdministratorsSelect, id).Scan(
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

func (o *administrators) ReadByEmail(email string) (*dao.Administrator, error) {
	administrator := &dao.Administrator{}
	err := o.db.QueryRow(queries.AdministratorsSelectByEmail, email).Scan(
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

func (o *administrators) Update(administrator *dao.Administrator) error {
	_, err := o.db.Exec(
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

// func (o *administrators) Delete(id string) (*dao.Administrator, error) {
// 	return nil, nil
// }

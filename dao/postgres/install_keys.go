package postgres

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
)

// InstallKeyHashType is a type alias representing the postgres ENUM for the key_hash_type column
type InstallKeyHashType string

const (
	// InstallKeyHashTypeSHA256 is a string for the sha256 key_hash_type ENUM
	InstallKeyHashTypeSHA256 = "SHA256"
)

type installKeys struct {
	db               *sql.DB
	installKeySecret string
}

// NewInstallKeys returns an instance implementing the Administrators interface
func NewInstallKeys(db *sql.DB, installKeySecret string) dao.InstallKeys {
	return &installKeys{db: db, installKeySecret: installKeySecret}
}

func (ik *installKeys) Create(administrator *dao.Administrator) (string, error) {
	if administrator == nil {
		return "", errors.New("invalid administrator")
	}
	if administrator.ID == "" {
		return "", errors.New("invalid administrator id")
	}
	if administrator.AuthorizationRole == "" {
		return "", errors.New("invalid authorization role")
	}
	claimsData := make(map[string]interface{})
	claimsData[psJWT.OrganizationIDKey] = administrator.OrganizationID
	claimsData[psJWT.AdministratorIDKey] = administrator.ID
	claimsData[psJWT.AuthorizationRoleKey] = administrator.AuthorizationRole
	claimsData[psJWT.TokenTypeKey] = psJWT.InstallKeySingleUse
	claimsData[psJWT.ClaimsTypeKey] = psJWT.InstallKey
	installKeyJWT, err := psJWT.GenerateJWT(ik.installKeySecret, psJWT.InstallKeySingleUse, psJWT.InstallKey, claimsData)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(installKeyJWT))
	installKeyHash := hex.EncodeToString(sum[:])
	var id string
	err = ik.db.QueryRow(queries.InstallKeysInsert, installKeyHash, InstallKeyHashTypeSHA256, administrator.ID, administrator.OrganizationID).Scan(&id)
	if err != nil {
		return "", err
	}
	return installKeyJWT, nil
}

func (ik *installKeys) Validate(installKeyJWT string) (*dao.InstallKey, error) {
	err := psJWT.DecodeJWT(ik.installKeySecret, installKeyJWT, psJWT.InstallKey)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256([]byte(installKeyJWT))
	installKeyHash := hex.EncodeToString(sum[:])

	var key dao.InstallKey
	err = ik.db.QueryRow(queries.InstallKeysSelect, installKeyHash).Scan(
		&key.AdministratorID,
		&key.OrganizationID,
	)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (ik *installKeys) Delete(installKeyJWT string) (int64, error) {
	sum := sha256.Sum256([]byte(installKeyJWT))
	installKeyHash := hex.EncodeToString(sum[:])

	result, err := ik.db.Exec(queries.InstallKeysDelete, installKeyHash)
	if err != nil {
		return 0, err
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsDeleted, nil
}

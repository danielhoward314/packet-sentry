package postgres

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
)

type organizations struct {
	db *sql.DB
}

// BillingPlanType is a type alias representing the postgres ENUM for the billing_plan_type column
type BillingPlanType string

const (
	// FreeBillingPlanType is a string for the free billing_plan_type ENUM
	FreeBillingPlanType = "FREE"
)

// NewOrganizations returns an instance implementing the Organizations interface
func NewOrganizations(db *sql.DB) dao.Organizations {
	return &organizations{db: db}
}

func (o *organizations) Create(organization *dao.Organization) (string, error) {
	if organization == nil {
		return "", errors.New("invalid organization")
	}
	if organization.PrimaryAdministratorEmail == "" {
		return "", errors.New("invalid organization primary administrator email")
	}
	if organization.Name == "" {
		return "", errors.New("invalid organization name")
	}
	var id string
	err := o.db.QueryRow(queries.OrganizationsInsert, organization.PrimaryAdministratorEmail, organization.Name, FreeBillingPlanType).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (o *organizations) Read(id string) (*dao.Organization, error) {
	organization := &dao.Organization{}
	err := o.db.QueryRow(queries.OrganizationsSelect, id).Scan(
		&organization.ID,
		&organization.PrimaryAdministratorEmail,
		&organization.Name,
		&organization.BillingPlanType,
	)
	if err != nil {
		return nil, err
	}
	return organization, nil
}

// func (o *organizations) Update(*dao.Organization) (*dao.Organization, error) {
// 	return nil, nil
// }

// func (o *organizations) Delete(id string) (*dao.Organization, error) {
// 	return nil, nil
// }

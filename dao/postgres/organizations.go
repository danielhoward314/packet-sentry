package postgres

import (
	"database/sql"
	"encoding/json"
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
	// TenDevicesBillingPlan is a string for the 10 devices billing_plan_type ENUM
	TenDevicesBillingPlan = "10_DEVICES_99_MONTH"
	// FiftyDevicesBillingPlan is a string for the 50 devices billing_plan_type ENUM
	FiftyDevicesBillingPlan = "50_DEVICES_399_MONTH"
	// OneHundreDevicesBillingPlan is a string for the 100 devices billing_plan_type ENUM
	OneHundreDevicesBillingPlan = "100_DEVICES_799_MONTH"
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
	err := o.db.QueryRow(
		queries.OrganizationsInsert,
		organization.PrimaryAdministratorEmail,
		organization.Name,
		TenDevicesBillingPlan,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (o *organizations) Read(id string) (*dao.Organization, error) {
	organization := &dao.Organization{}
	var paymentDetailsJSON []byte
	err := o.db.QueryRow(queries.OrganizationsSelect, id).Scan(
		&organization.ID,
		&organization.PrimaryAdministratorEmail,
		&organization.Name,
		&organization.BillingPlanType,
		&paymentDetailsJSON,
	)
	if err != nil {
		return nil, err
	}

	paymentDetails := &dao.PaymentDetails{}
	if len(paymentDetailsJSON) != 0 {
		err = json.Unmarshal(paymentDetailsJSON, paymentDetails)
		if err != nil {
			return nil, err
		}
	}
	organization.PaymentDetails = paymentDetails

	return organization, nil
}

func isBillingPlanTypeValid(plan string) bool {
	return (plan == TenDevicesBillingPlan ||
		plan == FiftyDevicesBillingPlan ||
		plan == OneHundreDevicesBillingPlan)
}

func (o *organizations) Update(organization *dao.Organization) error {
	if organization == nil {
		return errors.New("invalid organization")
	}
	if !isBillingPlanTypeValid(organization.BillingPlanType) {
		return errors.New("invalid billing plan type for update")
	}

	paymentDetailsJSON, err := json.Marshal(organization.PaymentDetails)
	if err != nil {
		return err
	}

	_, err = o.db.Exec(
		queries.OrganizationsUpdate,
		organization.Name,
		organization.BillingPlanType,
		paymentDetailsJSON,
		organization.ID,
	)

	if err != nil {
		return err
	}
	return nil
}

// func (o *organizations) Delete(id string) (*dao.Organization, error) {
// 	return nil, nil
// }

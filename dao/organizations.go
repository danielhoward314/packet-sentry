package dao

type Organization struct {
	ID                        string `json:"id"`
	PrimaryAdministratorEmail string `json:"primary_administrator_email"`
	Name                      string `json:"name"`
	BillingPlanType           string `json:"billing_plan_type"`
}

type Organizations interface {
	Create(organizations *Organization) (string, error)
	Read(id string) (*Organization, error)
	// Update(*Organization) (*Organization, error)
	// Delete(id string) (*Organization, error)
}

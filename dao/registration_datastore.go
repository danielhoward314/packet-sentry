package dao

type Registration struct {
	OrganizationID  string `json:"organization_id"`
	AdministratorID string `json:"administrator_id"`
	EmailCode       string `json:"email_code"`
}

// RegistrationDatastore defines the interface for registration operations in a key-value datastore
type RegistrationDatastore interface {
	Create(registration *Registration) (string, string, error)
	Read(token string) (*Registration, error)
	Delete(token string) error
}

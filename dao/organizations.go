package dao

type Organization struct {
	ID                        string          `json:"id"`
	PrimaryAdministratorEmail string          `json:"primary_administrator_email"`
	Name                      string          `json:"name"`
	BillingPlanType           string          `json:"billing_plan_type"`
	PaymentDetails            *PaymentDetails `json:"payment_details"`
}

type PaymentDetails struct {
	CardName        string `json:"cardName"`
	AddressLineOne  string `json:"addressLineOne"`
	AddressLineTwo  string `json:"addressLineTwo"`
	CardNumber      string `json:"cardNumber"`
	ExpirationMonth string `json:"expirationMonth"`
	ExpirationYear  string `json:"expirationYear"`
	Cvc             string `json:"cvc"`
}

type Organizations interface {
	Create(organizations *Organization) (string, error)
	Read(id string) (*Organization, error)
	Update(*Organization) error
	// Delete(id string) (*Organization, error)
}

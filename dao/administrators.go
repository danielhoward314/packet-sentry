package dao

type Administrator struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	DisplayName       string `json:"display_name"`
	PasswordHashType  string `json:"password_hash_type"`
	PasswordHash      string `json:"password_hash"`
	OrganizationID    string `json:"organization_id"`
	Verified          bool   `json:"verified"`
	AuthorizationRole string `json:"authorization_role"`
}

type Administrators interface {
	Create(administrator *Administrator, passwordCleartext string) (string, error)
	Read(id string) (*Administrator, error)
	ReadByEmail(email string) (*Administrator, error)
	Update(*Administrator) error
	// Delete(id string) (*Administrator, error)
}

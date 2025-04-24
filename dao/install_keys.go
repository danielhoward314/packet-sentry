package dao

type InstallKey struct {
	ID              string
	KeyHashType     string
	KeyHash         string
	AdministratorID string
	OrganizationID  string
}

type InstallKeys interface {
	Create(administrator *Administrator) (string, error)
	Validate(key string) (*InstallKey, error)
	Delete(key string) (int64, error)
}

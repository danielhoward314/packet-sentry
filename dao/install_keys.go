package dao

type InstallKey struct {
	ID              string `json:"id"`
	KeyHashType     string `json:"key_hash_type"`
	KeyHash         string `json:"key_hash"`
	AdministratorID string `json:"administrator_id"`
}
type InstallKeys interface {
	Create(administrator *Administrator) (string, error)
	ReadByKey(key string) (*InstallKey, error)
	DeleteByKey(key string) error
}

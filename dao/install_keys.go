package dao

type InstallKeys interface {
	Create(administrator *Administrator) (string, error)
	Validate(key string) error
	Delete(key string) (int64, error)
}

package dao

// Datastore exposes services that fulfill the primary datastore interfaces
type Datastore struct {
	Administrators Administrators
	InstallKeys    InstallKeys
	Organizations  Organizations
}

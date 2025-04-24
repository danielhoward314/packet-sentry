package dao

// Datastore exposes services that fulfill the primary datastore interfaces
type Datastore struct {
	Administrators Administrators
	Devices        Devices
	InstallKeys    InstallKeys
	Organizations  Organizations
}

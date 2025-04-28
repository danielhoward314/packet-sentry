package dao

type CaptureConfig struct {
	Bpf         string `json:"bpf"`
	DeviceName  string `json:"deviceName"`
	Promiscuous bool   `json:"promiscuous"`
	SnapLen     int32  `json:"snapLen"`
}

type Device struct {
	ID                       string
	OSUniqueIdentifier       string
	ClientCertPEM            string
	ClientCertFingerprint    string
	OrganizationID           string
	PCapVersion              string
	Interfaces               []string
	InterfaceBPFAssociations map[string]map[uint64]CaptureConfig
	PreviousAssociations     map[string]map[uint64]CaptureConfig
}

type Devices interface {
	Create(device *Device) error
	GetDeviceByPredicate(predicateName, predicateValue string) (*Device, error)
	List(organizationID string) ([]*Device, error)
	Update(device *Device) error
}

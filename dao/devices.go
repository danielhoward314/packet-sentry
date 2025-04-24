package dao

type BpfAssociation struct {
	Bpf         string `json:"bpf"`
	DeviceName  string `json:"deviceName"`
	Promiscuous bool   `json:"promiscuous"`
	SnapLen     int    `json:"snapLen"`
}

type Device struct {
	ID                       string
	OSUniqueIdentifier       string
	ClientCertPEM            string
	ClientCertFingerprint    string
	OrganizationID           string
	InterfaceBPFAssociations map[string]map[uint64]BpfAssociation
	PreviousAssociations     map[string]map[uint64]BpfAssociation
}

type Devices interface {
	Create(device *Device) error
	GetDeviceByOSUniqueIdentifier(osID string) (*Device, error)
	Update(device *Device) error
}

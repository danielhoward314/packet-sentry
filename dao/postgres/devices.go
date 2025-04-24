package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
)

type devices struct {
	db *sql.DB
}

func NewDevices(db *sql.DB) dao.Devices {
	return &devices{db: db}
}

func (d *devices) Create(device *dao.Device) error {
	if device == nil {
		return errors.New("invalid device")
	}
	if device.OSUniqueIdentifier == "" {
		return errors.New("invalid os_unique_identifier")
	}
	if device.ClientCertPEM == "" {
		return errors.New("invalid client_cert_pem")
	}
	if device.ClientCertFingerprint == "" {
		return errors.New("invalid client_cert_fingerprint")
	}
	if device.OrganizationID == "" {
		return errors.New("invalid organization_id")
	}
	return d.db.QueryRow(
		queries.DevicesInsert,
		device.OSUniqueIdentifier,
		device.ClientCertPEM,
		device.ClientCertFingerprint,
		device.OrganizationID,
	).Scan(&device.ID)
}

func (d *devices) GetDeviceByOSUniqueIdentifier(osID string) (*dao.Device, error) {
	if osID == "" {
		return nil, errors.New("invalid os unique identifier")
	}
	row := d.db.QueryRow(queries.DevicesSelectByOSUniqueIdentifier, osID)

	var device dao.Device
	var interfaceBPFJSON, previousBPFJSON []byte

	err := row.Scan(
		&device.ID,
		&device.OSUniqueIdentifier,
		&device.ClientCertPEM,
		&device.ClientCertFingerprint,
		&device.OrganizationID,
		&interfaceBPFJSON,
		&previousBPFJSON,
	)
	if err != nil {
		return nil, err
	}

	// The BPF hash key is stored as a string in the database, but we need to convert it to uint64
	device.InterfaceBPFAssociations, err = parseNestedJSONToUint64Map(interfaceBPFJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing interface_bpf_associations: %w", err)
	}
	device.PreviousAssociations, err = parseNestedJSONToUint64Map(previousBPFJSON)
	if err != nil {
		return nil, fmt.Errorf("parsing previous_associations: %w", err)
	}

	return &device, nil
}

func parseNestedJSONToUint64Map(input []byte) (map[string]map[uint64]dao.BpfAssociation, error) {
	var raw map[string]map[string]dao.BpfAssociation
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]map[uint64]dao.BpfAssociation)
	for iface, filters := range raw {
		result[iface] = make(map[uint64]dao.BpfAssociation)
		for strKey, assoc := range filters {
			uintKey, err := strconv.ParseUint(strKey, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse key '%s' as uint64: %w", strKey, err)
			}
			result[iface][uintKey] = assoc
		}
	}
	return result, nil
}

func (d *devices) Update(device *dao.Device) error {
	if device == nil {
		return errors.New("invalid device")
	}
	if device.ID == "" {
		return errors.New("invalid device ID")
	}
	if device.OSUniqueIdentifier == "" {
		return errors.New("invalid os_unique_identifier")
	}
	if device.ClientCertPEM == "" {
		return errors.New("invalid client_cert_pem")
	}
	if device.ClientCertFingerprint == "" {
		return errors.New("invalid client_cert_fingerprint")
	}
	if device.OrganizationID == "" {
		return errors.New("invalid organization_id")
	}

	var err error

	// if the update has no associations, set it to the empty map
	// which is the Go equivalent to the default '{}' for the jsonb columns
	// if the update has associations, convert the uint64 keys to string
	// which is how they're stored in the database
	var interfaceBPFAssociations map[string]map[string]dao.BpfAssociation
	if device.InterfaceBPFAssociations == nil {
		interfaceBPFAssociations = make(map[string]map[string]dao.BpfAssociation)
	} else {
		interfaceBPFAssociations, err = convertUint64MapToStringJSON(device.InterfaceBPFAssociations)
		if err != nil {
			return fmt.Errorf("converting interface_bpf_associations: %w", err)
		}
	}
	var previousInterfaceBPFAssociations map[string]map[string]dao.BpfAssociation
	if device.PreviousAssociations == nil {
		previousInterfaceBPFAssociations = make(map[string]map[string]dao.BpfAssociation)
	} else {
		previousInterfaceBPFAssociations, err = convertUint64MapToStringJSON(device.PreviousAssociations)
		if err != nil {
			return fmt.Errorf("converting previous_associations: %w", err)
		}
	}

	interfaceBPFJSON, err := json.Marshal(interfaceBPFAssociations)
	if err != nil {
		return fmt.Errorf("marshalling interface_bpf_associations: %w", err)
	}
	previousBPFJSON, err := json.Marshal(previousInterfaceBPFAssociations)
	if err != nil {
		return fmt.Errorf("marshalling previous_associations: %w", err)
	}

	_, err = d.db.Exec(
		queries.DevicesUpdate,
		device.ClientCertPEM,
		device.ClientCertFingerprint,
		interfaceBPFJSON,
		previousBPFJSON,
		device.ID,
	)
	return err
}

func convertUint64MapToStringJSON(input map[string]map[uint64]dao.BpfAssociation) (map[string]map[string]dao.BpfAssociation, error) {
	result := make(map[string]map[string]dao.BpfAssociation)

	for iface, filters := range input {
		result[iface] = make(map[string]dao.BpfAssociation)
		for uintKey, assoc := range filters {
			strKey := strconv.FormatUint(uintKey, 10)
			result[iface][strKey] = assoc
		}
	}

	return result, nil
}

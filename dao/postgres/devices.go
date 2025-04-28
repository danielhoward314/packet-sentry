package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
	"github.com/lib/pq"
)

const (
	PredicateID                 = "id"
	PredicateOSUniqueIdentifier = "os_unique_identifier"
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

func (d *devices) GetDeviceByPredicate(predicateName, predicateValue string) (*dao.Device, error) {
	if predicateName == "" {
		return nil, errors.New("empty predicate name for where clause")
	}
	if predicateValue == "" {
		return nil, errors.New("empty predicate value for where clause")
	}
	var row *sql.Row
	if predicateName == PredicateID {
		row = d.db.QueryRow(queries.DevicesSelectById, predicateValue)
	} else if predicateName == PredicateOSUniqueIdentifier {
		row = d.db.QueryRow(queries.DevicesSelectByOSUniqueIdentifier, predicateValue)
	} else {
		return nil, errors.New("invalid predicate name for where clause")
	}

	var device dao.Device
	var interfaces []string
	var interfaceBPFJSON, previousBPFJSON []byte

	err := row.Scan(
		&device.ID,
		&device.OSUniqueIdentifier,
		&device.ClientCertPEM,
		&device.ClientCertFingerprint,
		&device.OrganizationID,
		&device.PCapVersion,
		pq.Array(&interfaces),
		&interfaceBPFJSON,
		&previousBPFJSON,
	)
	if err != nil {
		return nil, err
	}

	device.Interfaces = interfaces

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
	// ensure we will always set `interfaces` column to TEXT[]
	if device.Interfaces == nil {
		device.Interfaces = make([]string, 0)
	}

	var err error

	// if the update has no associations, set it to the empty map
	// which is the Go equivalent to the default '{}' for the jsonb columns
	// if the update has associations, convert the uint64 keys to string
	// which is how they're stored in the database
	var interfaceBPFAssociations map[string]map[string]dao.CaptureConfig
	if device.InterfaceBPFAssociations == nil {
		interfaceBPFAssociations = make(map[string]map[string]dao.CaptureConfig)
	} else {
		interfaceBPFAssociations, err = convertUint64MapToStringJSON(device.InterfaceBPFAssociations)
		if err != nil {
			return fmt.Errorf("converting interface_bpf_associations: %w", err)
		}
	}
	var previousInterfaceBPFAssociations map[string]map[string]dao.CaptureConfig
	if device.PreviousAssociations == nil {
		previousInterfaceBPFAssociations = make(map[string]map[string]dao.CaptureConfig)
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
		device.PCapVersion,
		pq.Array(device.Interfaces),
		interfaceBPFJSON,
		previousBPFJSON,
		device.ID,
	)
	return err
}

func (d *devices) List(organizationID string) ([]*dao.Device, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("empty organization id")
	}

	var devices []*dao.Device
	rows, rowsErr := d.db.Query(queries.DevicesSelectByOrganizationID, organizationID)
	if rowsErr != nil {
		return nil, rowsErr
	}

	for rows.Next() {
		var device dao.Device
		var interfaces []string
		var interfaceBPFJSON, previousBPFJSON []byte

		rowErr := rows.Scan(
			&device.ID,
			&device.OSUniqueIdentifier,
			&device.ClientCertPEM,
			&device.ClientCertFingerprint,
			&device.OrganizationID,
			&device.PCapVersion,
			pq.Array(&interfaces),
			&interfaceBPFJSON,
			&previousBPFJSON,
		)
		if rowErr != nil {
			return nil, rowErr
		}

		device.Interfaces = interfaces

		// The BPF hash key is stored as a string in the database, but we need to convert it to uint64
		device.InterfaceBPFAssociations, rowErr = parseNestedJSONToUint64Map(interfaceBPFJSON)
		if rowErr != nil {
			return nil, fmt.Errorf("parsing interface_bpf_associations: %w", rowErr)
		}
		device.PreviousAssociations, rowErr = parseNestedJSONToUint64Map(previousBPFJSON)
		if rowErr != nil {
			return nil, fmt.Errorf("parsing previous_associations: %w", rowErr)
		}

		devices = append(devices, &device)
	}

	return devices, nil
}

func parseNestedJSONToUint64Map(input []byte) (map[string]map[uint64]dao.CaptureConfig, error) {
	var raw map[string]map[string]dao.CaptureConfig
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]map[uint64]dao.CaptureConfig)
	for iface, filters := range raw {
		result[iface] = make(map[uint64]dao.CaptureConfig)
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

func convertUint64MapToStringJSON(input map[string]map[uint64]dao.CaptureConfig) (map[string]map[string]dao.CaptureConfig, error) {
	result := make(map[string]map[string]dao.CaptureConfig)

	for iface, filters := range input {
		result[iface] = make(map[string]dao.CaptureConfig)
		for uintKey, assoc := range filters {
			strKey := strconv.FormatUint(uintKey, 10)
			result[iface][strKey] = assoc
		}
	}

	return result, nil
}

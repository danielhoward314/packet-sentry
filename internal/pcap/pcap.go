package pcap

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const (
	logAttrValSvcName = "pcapManager"
)

// PCapManager is the interface for managing packet capture for all interfaces and associated filters.
type PCapManager interface {
	EnsureReady() error
	StopOne(ifaceName string, filter string) error
	StartAll()
	StopAll()
}

// NewPCapManager returns a new implementation instance of the PCapManager interface.
func NewPCapManager(ctx context.Context, baseLogger *slog.Logger) PCapManager {
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))
	return &pcapManager{
		ctx:                            ctx,
		ifaceNameToFiltersAssociations: make(map[string]map[uint64]*packetCapture),
		interfaces:                     make(map[string]*pcap.Interface),
		logger:                         childLogger,
	}
}

// CaptureConfig holds configuration for a packet capture
type CaptureConfig struct {
	BPF         string        `json:"bpf"`
	DeviceName  string        `json:"deviceName"`
	Promiscuous bool          `json:"promiscuous"`
	SnapLen     int32         `json:"snapLen"`
	Timeout     time.Duration `json:"timeout"`
}

// LogValue implements the slog.LogValuer interface for the CaptureConfig struct
func (cc *CaptureConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String(psLog.KeyBPF, cc.BPF),
		slog.String(psLog.KeyDeviceName, cc.DeviceName),
		slog.Bool(psLog.KeyPromiscuous, cc.Promiscuous),
		slog.Int64(psLog.KeySnapLen, int64(cc.SnapLen)),
		slog.String(psLog.KeyTimeout, cc.Timeout.String()),
	)
}

type pcapManager struct {
	ifaceNameToFiltersAssociations map[string]map[uint64]*packetCapture
	ctx                            context.Context
	interfaces                     map[string]*pcap.Interface
	logger                         *slog.Logger
	mu                             sync.Mutex
	pcapVersion                    string
	wg                             sync.WaitGroup
}

// EnsureReady confirms pcap is ready for use and reports all available interfaces to the backend.
func (m *pcapManager) EnsureReady() error {
	m.logger.With(psLog.KeyFunction, "PCapManager.EnsureReady")

	m.pcapVersion = pcap.Version()
	m.logger.Info("collected pcap version", slog.String(psLog.KeyPCapVersion, m.pcapVersion))
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		m.logger.Error("failed to find all devices with pcap", psLog.KeyError, err)
		return err
	}

	for _, iface := range interfaces {
		m.logger.Info("found device", slog.String(psLog.KeyDeviceName, iface.Name))
		m.ifaceNameToFiltersAssociations[iface.Name] = make(map[uint64]*packetCapture)
		m.interfaces[iface.Name] = &iface
	}
	// TODO: send interfaces to backend to drive UI for writing BPF for a given interface
	return nil
}

// StopOne stops and removes a capture for the given interface name for the given filter
func (m *pcapManager) StopOne(ifaceName string, filter string) error {
	m.logger.With(psLog.KeyFunction, "PCapManager.StopOne")
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, exists := m.ifaceNameToFiltersAssociations[ifaceName]
	if !exists {
		m.logger.Error("failed to find interface by name", slog.String(psLog.KeyDeviceName, ifaceName))
		return fmt.Errorf("no interface found by name '%s'", ifaceName)
	}

	filterHash := xxhash.Sum64([]byte(filter))
	capture, exists := iface[filterHash]
	if !exists {
		m.logger.Error(
			"failed to find capture by hash of bpf",
			slog.String(psLog.KeyBPF, filter),
			slog.Uint64(psLog.KeyBPFHash, filterHash),
		)
		return fmt.Errorf("no capture found with filter '%s'", filter)
	}
	capture.Stop()
	m.logger.Info(
		"deleting bpf hash entry from interface name's map",
		slog.String(psLog.KeyDeviceName, ifaceName),
		slog.Uint64(psLog.KeyBPFHash, filterHash),
	)
	delete(m.ifaceNameToFiltersAssociations[ifaceName], filterHash)
	return nil
}

// StartAll starts all packet captures for all interfaces and associated filters.
func (m *pcapManager) StartAll() {
	m.logger.With(psLog.KeyFunction, "PCapManager.StartAll")

	// TODO: fetch filters from backend rather than hardcoding
	hardcodedFilter := "tcp and port 3000"
	filterHash := xxhash.Sum64([]byte(hardcodedFilter))
	// TODO: replace this hard-coding below that uses the same filter to all interfaces
	// with configuration per iface that comes from the backend
	for ifaceName := range m.ifaceNameToFiltersAssociations {
		capture, err := newpacketCapture(
			m.ctx,
			m.logger,
			&CaptureConfig{
				BPF:         hardcodedFilter,
				DeviceName:  ifaceName,
				Promiscuous: false,
				SnapLen:     1600,
				Timeout:     pcap.BlockForever,
			},
			&m.wg,
		)
		if err != nil {
			m.logger.Error("failed to create packet capture", psLog.KeyError, err)
			panic("failed to create packet capture")
		}
		if len(m.ifaceNameToFiltersAssociations[ifaceName]) == 0 {
			m.logger.Info(
				"initializing filter associations map for interface",
				slog.String(psLog.KeyDeviceName, ifaceName),
			)
			m.ifaceNameToFiltersAssociations[ifaceName] = make(map[uint64]*packetCapture)
		}
		m.logger.Info(
			"associating filter to interface",
			slog.String(psLog.KeyDeviceName, ifaceName),
			slog.String(psLog.KeyBPF, hardcodedFilter),
			slog.Uint64(psLog.KeyBPFHash, filterHash),
		)
		m.ifaceNameToFiltersAssociations[ifaceName][filterHash] = capture
	}
	// TODO: replace this hard-coding above that uses the same filter to all interfaces
	// with configuration per iface that comes from the backend

	for ifaceNameKey, filtersForIFace := range m.ifaceNameToFiltersAssociations {
		for filterHashKey, packetCapture := range filtersForIFace {
			m.logger.Info(
				"starting packet capture",
				slog.String(psLog.KeyDeviceName, ifaceNameKey),
				slog.String(psLog.KeyBPF, packetCapture.config.BPF),
				slog.Uint64(psLog.KeyBPFHash, filterHashKey),
			)
			err2 := packetCapture.Start()
			if err2 != nil {
				m.logger.Error(
					"failed to start packet capture",
					slog.String(psLog.KeyDeviceName, ifaceNameKey),
					slog.String(psLog.KeyBPF, packetCapture.config.BPF),
					slog.Uint64(psLog.KeyBPFHash, filterHashKey),
				)
				continue
			}
		}
	}
}

// StartAll stops all packet captures for all interfaces and associated filters.
func (m *pcapManager) StopAll() {
	m.logger.With(psLog.KeyFunction, "PCapManager.StopAll")

	m.mu.Lock()
	defer m.mu.Unlock()

	for ifaceNameKey, filtersForIFace := range m.ifaceNameToFiltersAssociations {
		for filterHashKey, packetCapture := range filtersForIFace {
			m.logger.Info(
				"stopping packet capture",
				slog.String(psLog.KeyDeviceName, ifaceNameKey),
				slog.String(psLog.KeyBPF, packetCapture.config.BPF),
				slog.Uint64(psLog.KeyBPFHash, filterHashKey),
			)
			packetCapture.Stop()
		}
	}

	// reset the map
	m.ifaceNameToFiltersAssociations = make(map[string]map[uint64]*packetCapture)
}

// packetCapture holds the config used to create a capture and the handle to the live capture
type packetCapture struct {
	cancelFunc context.CancelFunc
	config     *CaptureConfig
	ctx        context.Context
	handle     *pcap.Handle
	logger     *slog.Logger
	wg         *sync.WaitGroup
}

func newpacketCapture(parentCtx context.Context, parentLogger *slog.Logger, config *CaptureConfig, wg *sync.WaitGroup) (*packetCapture, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	childLogger := parentLogger.With(psLog.KeyCaptureConfig, config)

	return &packetCapture{
		config:     config,
		ctx:        ctx,
		cancelFunc: cancel,
		logger:     childLogger,
		wg:         wg,
	}, nil
}

func (pc *packetCapture) cleanup() {
	pc.logger.With(psLog.KeyFunction, "packetCapture.cleanup")
	if pc.handle != nil {
		pc.logger.Info("closing packet capture handle")
		pc.handle.Close()
		pc.handle = nil
	} else {
		pc.logger.Info("cleanup did not find handle for packet capture")
	}
	pc.wg.Done()
}

// Start begins the packet capture process
func (pc *packetCapture) Start() error {
	pc.logger.With(psLog.KeyFunction, "packetCapture.Start")
	var err error
	pc.wg.Add(1)

	pc.logger.Info("opening live packet capture")
	pc.handle, err = pcap.OpenLive(
		pc.config.DeviceName,
		pc.config.SnapLen,
		pc.config.Promiscuous,
		pc.config.Timeout,
	)
	if err != nil {
		pc.wg.Done()
		pc.logger.Error("error opening device", psLog.KeyError, err)
		return err
	}

	pc.logger.Info("setting BPF")
	err = pc.handle.SetBPFFilter(pc.config.BPF)
	if err != nil {
		pc.logger.Error("error setting BPF", psLog.KeyError, err)
		pc.cleanup()
		return err
	}

	go func() {
		defer pc.cleanup()
		packetSource := gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
		packetChan := packetSource.Packets()
		pc.logger.Info("Packet source created, entering capture loop")

		for {
			select {
			case packet, ok := <-packetChan:
				if !ok {
					pc.logger.Info("Packet channel closed")
					return
				}
				// TODO: send data to backend rather than log it
				pc.logger.Info("--------------packet start----------------")
				pc.logger.Info(packet.String())
				pc.logger.Info("--------------packet end------------------")

			case <-pc.ctx.Done():
				pc.logger.Info("context canceled, stopping packet capture")
				return
			}
		}
	}()

	pc.logger.Info("Packet capture started successfully")
	return nil
}

// Stop terminates the packet capture process
func (pc *packetCapture) Stop() {
	pc.logger.With(psLog.KeyFunction, "packetCapture.Stop")
	pc.logger.Info("calling packet capture's cancel func")
	pc.cancelFunc()
}

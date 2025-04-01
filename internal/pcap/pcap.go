package pcap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// PCapManager is the interface for managing packet capture for all interfaces and associated filters.
type PCapManager interface {
	EnsureReady() error
	RemoveCapture(ifaceName string, filter string) error
	StartAll()
	StopAll()
}

// NewPCapManager returns a new implementation instance of the PCapManager interface.
func NewPCapManager(ctx context.Context) PCapManager {
	return &pCapManager{
		ctx:                            ctx,
		ifaceNameToFiltersAssociations: make(map[string]map[uint64]*PacketCapture),
		interfaces:                     make(map[string]*pcap.Interface),
	}
}

type pCapManager struct {
	ifaceNameToFiltersAssociations map[string]map[uint64]*PacketCapture
	ctx                            context.Context
	interfaces                     map[string]*pcap.Interface
	mu                             sync.Mutex
	pcapVersion                    string
	wg                             sync.WaitGroup
}

// EnsureReady confirms pcap is ready for use and reports all available interfaces to the backend.
func (m *pCapManager) EnsureReady() error {
	m.pcapVersion = pcap.Version()
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		m.ifaceNameToFiltersAssociations[iface.Name] = make(map[uint64]*PacketCapture)
		m.interfaces[iface.Name] = &iface
	}
	// TODO: send interfaces to backend to drive UI for writing BPF for a given interface
	return nil
}

// RemoveCapture stops and removes a capture for the given interface name for the given filter
func (m *pCapManager) RemoveCapture(ifaceName string, filter string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, exists := m.ifaceNameToFiltersAssociations[ifaceName]
	if !exists {
		return fmt.Errorf("no interface found by name '%s'", ifaceName)
	}

	filterHash := xxhash.Sum64([]byte(filter))
	capture, exists := iface[filterHash]
	if !exists {
		return fmt.Errorf("no capture found with filter '%s'", filter)
	}

	capture.Stop()
	delete(m.ifaceNameToFiltersAssociations[ifaceName], filterHash)
	return nil
}

// StartAll starts all packet captures for all interfaces and associated filters.
func (m *pCapManager) StartAll() {
	// TODO: fetch filters from backend rather than hardcoding
	hardcodedFilter := "tcp and port 3000"
	filterHash := xxhash.Sum64([]byte(hardcodedFilter))
	iface := "lo0"
	capture, err := newPacketCapture(
		m.ctx,
		CaptureConfig{
			bpfFilter:   hardcodedFilter,
			deviceName:  iface,
			promiscuous: false,
			snapLen:     1600,
			timeout:     pcap.BlockForever,
		},
		&m.wg,
	)
	if err != nil {
		panic("failed to create packet capture")
	}
	m.ifaceNameToFiltersAssociations[iface][filterHash] = capture

	for _, filtersForIFace := range m.ifaceNameToFiltersAssociations {
		for _, packetCapture := range filtersForIFace {
			err2 := packetCapture.Start()
			if err2 != nil {
				// TODO: report error to backend and/or log file
				continue
			}
		}
	}
}

// StartAll stops all packet captures for all interfaces and associated filters.
func (m *pCapManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, filtersForIFace := range m.ifaceNameToFiltersAssociations {
		for _, packetCapture := range filtersForIFace {
			packetCapture.Stop()
		}
	}

	// reset the map
	m.ifaceNameToFiltersAssociations = make(map[string]map[uint64]*PacketCapture)
}

// PacketCapture holds the config used to create a capture and the handle to the live capture
type PacketCapture struct {
	cancelFunc context.CancelFunc
	config     CaptureConfig
	ctx        context.Context
	handle     *pcap.Handle
	wg         *sync.WaitGroup
}

// CaptureConfig holds configuration for a packet capture
type CaptureConfig struct {
	bpfFilter   string
	deviceName  string
	promiscuous bool
	snapLen     int32
	timeout     time.Duration
}

func newPacketCapture(parentCtx context.Context, config CaptureConfig, wg *sync.WaitGroup) (*PacketCapture, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	return &PacketCapture{
		config:     config,
		ctx:        ctx,
		cancelFunc: cancel,
		wg:         wg,
	}, nil
}

func (pc *PacketCapture) cleanup() {
	if pc.handle != nil {
		pc.handle.Close()
		pc.handle = nil
	}
	pc.wg.Done()
}

// Start begins the packet capture process
func (pc *PacketCapture) Start() error {
	var err error
	pc.wg.Add(1)

	pc.handle, err = pcap.OpenLive(
		pc.config.deviceName,
		pc.config.snapLen,
		pc.config.promiscuous,
		pc.config.timeout,
	)
	if err != nil {
		pc.wg.Done()
		return fmt.Errorf("error opening device %s: %v", pc.config.deviceName, err)
	}

	if err := pc.handle.SetBPFFilter(pc.config.bpfFilter); err != nil {
		pc.cleanup()
		return fmt.Errorf("error setting BPF filter %s: %v", pc.config.bpfFilter, err)
	}

	go func() {
		defer pc.cleanup()

		packetSource := gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
		packetChan := packetSource.Packets()

		for {
			select {
			case packet, ok := <-packetChan:
				if !ok {
					return
				}
				fmt.Printf("[Filter: %s] --------------packet start----------------\n", pc.config.bpfFilter)
				fmt.Println(packet)
				fmt.Printf("[Filter: %s] --------------packet end------------------\n", pc.config.bpfFilter)

			case <-pc.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop terminates the packet capture process
func (pc *PacketCapture) Stop() {
	pc.cancelFunc()
}

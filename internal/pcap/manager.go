package pcap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
)

const (
	logAttrValSvcName = "pcapManager"
)

// PCapManager is the interface for managing packet capture for all interfaces and associated filters.
type PCapManager interface {
	StartAll()
	StopAll()
	StopOne(ifaceName string, filterHash uint64, filter string) error
}

type pcapManager struct {
	cancelFunc                     context.CancelFunc
	client                         *http.Client
	clientMu                       sync.RWMutex
	command                        *broadcast.Command
	commandMu                      sync.RWMutex
	ctx                            context.Context
	ifaceNameToFiltersAssociations map[string]map[uint64]*packetCapture
	interfaces                     map[string]*pcap.Interface
	logger                         *slog.Logger
	mTLSClientBroadcaster          *broadcast.MTLSClientBroadcaster
	mu                             sync.Mutex
	packetChan                     chan gopacket.Packet
	commandsBroadcaster            *broadcast.CommandsBroadcaster
	pcapVersion                    string
	stopOnce                       sync.Once
	wg                             sync.WaitGroup
}

// NewPCapManager returns a new implementation instance of the PCapManager interface.
// The method also subscribes to broadcasts and starts the packet uploader goroutine.
func NewPCapManager(
	ctx context.Context,
	baseLogger *slog.Logger,
	commandsBroadcaster *broadcast.CommandsBroadcaster,
	mTLSClientBroadcaster *broadcast.MTLSClientBroadcaster,
) PCapManager {
	childCtx, cancelFunc := context.WithCancel(ctx)
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))

	return &pcapManager{
		cancelFunc:                     cancelFunc,
		commandsBroadcaster:            commandsBroadcaster,
		ctx:                            childCtx,
		ifaceNameToFiltersAssociations: make(map[string]map[uint64]*packetCapture),
		interfaces:                     make(map[string]*pcap.Interface),
		logger:                         childLogger,
		mTLSClientBroadcaster:          mTLSClientBroadcaster,
		packetChan:                     make(chan gopacket.Packet, 500),
	}
}

// StartAll coordinates packet capture with several key functions:
// (1) subscribes to commands to trigger fetching config
// (2) upon receiving `get_bpf_config` command, fetches config from the server
// (3) enforces config by starting all packet captures for all interfaces and associated filters
// (4) subscribes to mTLS client updates
func (m *pcapManager) StartAll() {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.StartAll")

	if m.mTLSClientBroadcaster == nil {
		fmt.Printf("client broadcaster is nil")
		return
	}
	if m.commandsBroadcaster == nil {
		fmt.Printf("commands broadcaster is nil")
		return
	}

	clientSubscription := m.mTLSClientBroadcaster.Subscribe()
	commandsSubscription := m.commandsBroadcaster.Subscribe()

	for {
		select {
		case clientUpdate := <-clientSubscription:
			m.clientMu.Lock()
			m.client = clientUpdate.Client
			m.clientMu.Unlock()
		case command := <-commandsSubscription:
			m.commandMu.Lock()
			commandName := command.Name
			m.commandMu.Unlock()
			switch commandName {
			case broadcast.CommandSendInterfaces:
				logger.Info("processing command", psLog.KeyCommand, broadcast.CommandSendInterfaces)
				err := m.sendInterfaces()
				if err != nil {
					logger.Error("failed to send interfaces to server", psLog.KeyError, err)
					continue
				}
			case broadcast.CommandGetBPFConfig:
				logger.Info("processing command", psLog.KeyCommand, broadcast.CommandGetBPFConfig)
				bpfConfig, err := m.fetchBPFConfig()
				if err != nil {
					logger.Error("failed to fetch BPF config", psLog.KeyError, err)
					continue
				}
				err = m.enforceConfig(bpfConfig)
				if err != nil {
					logger.Error("failed to enforce BPF config", psLog.KeyError, err)
					continue
				}
			default:
				// do nothing, command not for this manager
			}
		case pkt := <-m.packetChan:
			err := m.sendPacketEvent(pkt)
			if err != nil {
				logger.Error("failed to send packet event", psLog.KeyError, err)
				continue
			}
		case <-m.ctx.Done():
			logger.Error("pcap manager context canceled")
			return
		}
	}
}

// StopOne stops and removes a capture for the given interface name for the given filter hash
func (m *pcapManager) StopOne(ifaceName string, filterHash uint64, filter string) error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.StopOne")
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, exists := m.ifaceNameToFiltersAssociations[ifaceName]
	if !exists {
		logger.Error(
			"failed to find interface by name",
			slog.String(psLog.KeyDeviceName, ifaceName),
			slog.String(psLog.KeyBPF, filter),
			slog.Uint64(psLog.KeyBPFHash, filterHash),
		)
		return &PacketCaptureNotFound{
			FilterHash: filterHash,
			Filter:     filter,
		}
	}

	capture, exists := iface[filterHash]
	if !exists {
		logger.Error(
			"failed to find capture for interface by hash of bpf",
			slog.String(psLog.KeyDeviceName, ifaceName),
			slog.String(psLog.KeyBPF, filter),
			slog.Uint64(psLog.KeyBPFHash, filterHash),
		)
		return &PacketCaptureNotFound{
			FilterHash: filterHash,
			Filter:     filter,
		}
	}
	capture.Stop()
	logger.Info(
		"deleting bpf hash entry from interface name's map",
		slog.String(psLog.KeyDeviceName, ifaceName),
		slog.String(psLog.KeyBPF, filter),
		slog.Uint64(psLog.KeyBPFHash, filterHash),
	)
	delete(m.ifaceNameToFiltersAssociations[ifaceName], filterHash)
	return nil
}

// StopAll stops all packet captures for all interfaces and associated filters.
func (m *pcapManager) StopAll() {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.StopAll")

	logger.Info("stopping pcap manager and stopping any live captures")

	m.mu.Lock()
	defer m.mu.Unlock()

	for ifaceNameKey, filtersForIFace := range m.ifaceNameToFiltersAssociations {
		for filterHashKey, packetCapture := range filtersForIFace {
			logger.Info(
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

func (m *pcapManager) sendInterfaces() error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.sendInterfaces")

	m.pcapVersion = pcap.Version()
	logger.Info("collected pcap version", slog.String(psLog.KeyPCapVersion, m.pcapVersion))
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		logger.Error("failed to find all devices with pcap", psLog.KeyError, err)
		return err
	}

	reportRequest := &ReportInterfacesRequest{
		Interfaces:  make([]*InterfaceDetails, 0),
		PCapVersion: m.pcapVersion,
	}

	for _, iface := range interfaces {
		logger.Info("found device", slog.String(psLog.KeyDeviceName, iface.Name))
		m.ifaceNameToFiltersAssociations[iface.Name] = make(map[uint64]*packetCapture)
		m.interfaces[iface.Name] = &iface
		reportRequest.Interfaces = append(reportRequest.Interfaces, &InterfaceDetails{Name: iface.Name})
	}

	logger.Info("marshaling request body")
	bodyJSON, err := json.Marshal(reportRequest)
	if err != nil {
		logger.Error("failed to marshal interfaces report", psLog.KeyError, err)
		return err
	}

	logger.Info("creating http request")
	req, err := http.NewRequestWithContext(m.ctx, "POST", "https://localhost:9443/interfaces", bytes.NewBuffer(bodyJSON))
	if err != nil {
		logger.Error("failed to create interfaces report http request", psLog.KeyError, err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	logger.Info("sending http request", psLog.KeyURI, "/interfaces")
	m.clientMu.RLock()
	client := m.client
	m.clientMu.RUnlock()

	if client == nil {
		logger.Warn("no HTTPS client available")
		return fmt.Errorf("no HTTPS client available")
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send interface report", psLog.KeyError, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("server returned non-200")
		return fmt.Errorf("server returned non-200")
	}
	return nil
}

func (m *pcapManager) fetchBPFConfig() (*BPFResponse, error) {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.fetchBPFConfig")

	req, err := http.NewRequestWithContext(m.ctx, http.MethodGet, "https://localhost:9443/bpf", nil)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	logger.Info("sending http request", psLog.KeyURI, "/bpf")
	m.clientMu.RLock()
	client := m.client
	m.clientMu.RUnlock()
	if client == nil {
		logger.Error("no mTLS client available, cannot get BPF config")
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		logger.Error("failed to poll server", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("received non-200 response from /bpf", psLog.KeyError, resp.Status)
		return nil, err
	}

	logger.Info("reading http response body")
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body", psLog.KeyError, err)
		return nil, err
	}

	logger.Info("unmarshaling http response body")
	var bpfResponse BPFResponse
	err = json.Unmarshal(resBody, &bpfResponse)
	if err != nil {
		logger.Error("failed to unmarshal response body", psLog.KeyError, err)
		return nil, err
	}
	return &bpfResponse, nil
}

func (m *pcapManager) enforceConfig(bpfConfig *BPFResponse) error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.fetchBPFConfig")

	var errs []error

	if len(bpfConfig.Delete) > 0 {
		for ifaceName, bpfAssociationsToDelete := range bpfConfig.Delete {
			for filterHash, captureCfg := range bpfAssociationsToDelete {
				deleteErr := m.StopOne(ifaceName, filterHash, captureCfg.BPF)
				if deleteErr != nil {
					switch deleteErr.(type) {
					case *PacketCaptureNotFound:
						logger.Warn(
							"BPF association to delete is not found in associations for interface, skipping delete for this association",
							psLog.KeyError,
							deleteErr.Error(),
						)
						// don't append to errs, since a request to delete a BPF w/o a current live capture amounts to a no-op
						continue
					default:
						logger.Error(
							"failed to delete BPF association",
							slog.String(psLog.KeyDeviceName, ifaceName),
							slog.String(psLog.KeyBPF, captureCfg.BPF),
							slog.Uint64(psLog.KeyBPFHash, filterHash),
							psLog.KeyError,
							deleteErr.Error(),
						)
						errs = append(errs, deleteErr)
						continue
					}
				}
			}
		}
	}

	if len(bpfConfig.Update) > 0 {
		for ifaceName, bpfAssociationsToUpdate := range bpfConfig.Update {
			for filterHash, captureCfg := range bpfAssociationsToUpdate {
				updateErr := m.StopOne(ifaceName, filterHash, captureCfg.BPF)
				if updateErr != nil {
					switch updateErr.(type) {
					case *PacketCaptureNotFound:
						logger.Warn(
							"BPF association to update is not found in associations for interface, proceeding as if creating new filter association",
							psLog.KeyError,
							updateErr.Error(),
						)
						// don't append to errs, since a request to update a BPF w/o a current live capture amounts to a create request
					default:
						logger.Error(
							"failed to stop packet capture for BPF association before update, skipping update for this filter",
							slog.String(psLog.KeyDeviceName, ifaceName),
							slog.String(psLog.KeyBPF, captureCfg.BPF),
							slog.Uint64(psLog.KeyBPFHash, filterHash),
							psLog.KeyError,
							updateErr.Error(),
						)
						errs = append(errs, updateErr)
						continue
					}
				}

				updatedPacketCapture, updateErr := newPacketCapture(
					m.ctx,
					m.logger,
					&CaptureConfig{
						BPF:         captureCfg.BPF,
						DeviceName:  ifaceName,
						Promiscuous: captureCfg.Promiscuous,
						SnapLen:     captureCfg.SnapLen,
						Timeout:     pcap.BlockForever.Abs(),
					},
					&m.wg,
					m.packetChan,
				)
				if updateErr != nil {
					logger.Error(
						"failed to update packet capture for BPF association",
						slog.String(psLog.KeyDeviceName, ifaceName),
						slog.String(psLog.KeyBPF, captureCfg.BPF),
						slog.Uint64(psLog.KeyBPFHash, filterHash),
						psLog.KeyError,
						updateErr.Error(),
					)
					errs = append(errs, updateErr)
					continue
				}
				if m.ifaceNameToFiltersAssociations[ifaceName] == nil {
					m.ifaceNameToFiltersAssociations[ifaceName] = make(map[uint64]*packetCapture)
				}
				m.ifaceNameToFiltersAssociations[ifaceName][filterHash] = updatedPacketCapture
			}
		}
	}

	if len(bpfConfig.Create) > 0 {
		for ifaceName, bpfAssociationsToCreate := range bpfConfig.Create {
			for filterHash, captureCfg := range bpfAssociationsToCreate {
				createdPacketCapture, createErr := newPacketCapture(
					m.ctx,
					m.logger,
					&CaptureConfig{
						BPF:         captureCfg.BPF,
						DeviceName:  ifaceName,
						Promiscuous: captureCfg.Promiscuous,
						SnapLen:     captureCfg.SnapLen,
						Timeout:     pcap.BlockForever.Abs(),
					},
					&m.wg,
					m.packetChan,
				)
				if createErr != nil {
					logger.Error(
						"failed to create packet capture for BPF association",
						slog.String(psLog.KeyDeviceName, ifaceName),
						slog.String(psLog.KeyBPF, captureCfg.BPF),
						slog.Uint64(psLog.KeyBPFHash, filterHash),
						psLog.KeyError,
						createErr.Error(),
					)
					errs = append(errs, createErr)
					continue
				}
				if m.ifaceNameToFiltersAssociations[ifaceName] == nil {
					m.ifaceNameToFiltersAssociations[ifaceName] = make(map[uint64]*packetCapture)
				}
				m.ifaceNameToFiltersAssociations[ifaceName][filterHash] = createdPacketCapture
			}
		}
	}

	for _, bpfAssociations := range m.ifaceNameToFiltersAssociations {
		for _, packetCaptureToStart := range bpfAssociations {
			pcapStartErr := packetCaptureToStart.Start()
			if pcapStartErr != nil {
				errs = append(errs, pcapStartErr)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (m *pcapManager) sendPacketEvent(pkt gopacket.Packet) error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.sendPacketEvent")

	err := pkt.ErrorLayer()
	if err != nil {
		logger.Error("failed to decode packet", psLog.KeyError, err)
	}

	m.clientMu.RLock()
	client := m.client
	m.clientMu.RUnlock()

	if client == nil {
		// TODO: replace with a dropped packet log that gets sent, then wiped, on a 24-hour interval
		logger.Warn("no HTTPS client available, dropping packet", psLog.KeyDroppedPacket, pkt.String())
		return nil
	}

	// TODO: double-check whether this nesting of goroutines is okay
	// the idea is to spin up a goroutine that we won't wait for. if packet is dropped, it'll just log to the dropped packet file
	go func(pkt gopacket.Packet) {
		// TODO: replace https request with NATS/Kafka or similar high throughput ingestion
		logger.Info("marshaling request body")
		// TODO: replace with more refined request shape
		packetData := &PacketEvent{
			Data: pkt.Dump(),
		}
		bodyJSON, err := json.Marshal(packetData)
		if err != nil {
			// TODO: replace with a dropped packet log that gets sent, then wiped, on a 24-hour interval
			logger.Error("failed to marshal packet for upload", psLog.KeyError, err)
			return
		}

		logger.Info("creating http request")
		req, err := http.NewRequestWithContext(m.ctx, "POST", "https://localhost:9443/packets", bytes.NewBuffer(bodyJSON))
		if err != nil {
			// TODO: replace with a dropped packet log that gets sent, then wiped, on a 24-hour interval
			logger.Error("failed to create packet upload http request", psLog.KeyError, err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		logger.Info("sending http request", psLog.KeyURI, "/packets")
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("failed to upload packet", psLog.KeyError, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Error("server returned non-200")
			return
		}
	}(pkt)

	return nil
}

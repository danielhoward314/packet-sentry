package pcap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/gopacket/pcap"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
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
	agentMTLSClient                pbAgent.AgentServiceClient
	agentMTLSClientBroadcaster     *broadcast.AgentMTLSClientBroadcaster
	agentMTLSClientMu              sync.RWMutex
	cancelFunc                     context.CancelFunc
	commandsBroadcaster            *broadcast.CommandsBroadcaster
	commandMu                      sync.RWMutex
	currentStreamCancel            context.CancelFunc
	ctx                            context.Context
	ifaceNameToFiltersAssociations map[string]map[uint64]*packetCapture
	interfaces                     map[string]*pcap.Interface
	logger                         *slog.Logger
	mu                             sync.Mutex
	packetChan                     chan WrappedPacket
	packetStreamClient             pbAgent.AgentService_SendPacketEventClient
	pcapVersion                    string
	stopOnce                       sync.Once
	streamMu                       sync.Mutex
	wg                             sync.WaitGroup
}

// NewPCapManager returns a new implementation instance of the PCapManager interface.
func NewPCapManager(
	ctx context.Context,
	baseLogger *slog.Logger,
	commandsBroadcaster *broadcast.CommandsBroadcaster,
	agentMTLSClientBroadcaster *broadcast.AgentMTLSClientBroadcaster,
) PCapManager {
	childCtx, cancelFunc := context.WithCancel(ctx)
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))

	return &pcapManager{
		agentMTLSClientBroadcaster:     agentMTLSClientBroadcaster,
		cancelFunc:                     cancelFunc,
		commandsBroadcaster:            commandsBroadcaster,
		ctx:                            childCtx,
		ifaceNameToFiltersAssociations: make(map[string]map[uint64]*packetCapture),
		interfaces:                     make(map[string]*pcap.Interface),
		logger:                         childLogger,
		packetChan:                     make(chan WrappedPacket, 500),
	}
}

// StartAll coordinates packet capture with several key functions:
// (1) subscribes to commands to trigger fetching config
// (2) upon receiving `get_bpf_config` command, fetches config from the server
// (3) enforces config by starting all packet captures for all interfaces and associated filters
// (4) subscribes to mTLS client updates
func (m *pcapManager) StartAll() {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.StartAll")

	if m.agentMTLSClientBroadcaster == nil {
		fmt.Printf("client broadcaster is nil")
		return
	}
	if m.commandsBroadcaster == nil {
		fmt.Printf("commands broadcaster is nil")
		return
	}

	clientSubscription := m.agentMTLSClientBroadcaster.Subscribe()
	commandsSubscription := m.commandsBroadcaster.Subscribe()

	for {
		select {
		case clientUpdate := <-clientSubscription:
			m.agentMTLSClientMu.Lock()
			client := pbAgent.NewAgentServiceClient(clientUpdate.ClientConn)
			m.agentMTLSClient = client
			m.agentMTLSClientMu.Unlock()

			// If there's an existing stream, cancel it (which should cause CloseSend)
			m.streamMu.Lock()
			if m.currentStreamCancel != nil {
				m.currentStreamCancel()
			}
			// Create a new context derived from m.ctx for this stream
			ctx, cancel := context.WithCancel(m.ctx)
			// The stream needs its own context and cancel func so that it is distinct each time
			// and can be canceled any time a client update is received on this channel
			stream, err := client.SendPacketEvent(ctx)
			if err != nil {
				logger.Error("failed to open packet stream", psLog.KeyError, err)
				m.currentStreamCancel = cancel
				m.streamMu.Unlock()
				continue
			}
			m.packetStreamClient = stream
			m.currentStreamCancel = cancel
			m.streamMu.Unlock()
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

	m.stopOnce.Do(func() {
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
		m.cancelFunc()
	})
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

	reportRequest := &pbAgent.ReportInterfacesRequest{
		Interfaces:  make([]*pbAgent.InterfaceDetails, 0),
		PcapVersion: m.pcapVersion,
	}

	for _, iface := range interfaces {
		logger.Info("found device", slog.String(psLog.KeyDeviceName, iface.Name))
		m.ifaceNameToFiltersAssociations[iface.Name] = make(map[uint64]*packetCapture)
		m.interfaces[iface.Name] = &iface
		reportRequest.Interfaces = append(reportRequest.Interfaces, &pbAgent.InterfaceDetails{Name: iface.Name})
	}

	logger.Info("sending interfaces")
	m.agentMTLSClientMu.RLock()
	client := m.agentMTLSClient
	m.agentMTLSClientMu.RUnlock()

	if client == nil {
		logger.Warn("no agent gRPC client available")
		return fmt.Errorf("no agent gRPC client available")
	}

	_, err = client.ReportInterfaces(m.ctx, reportRequest)
	return err
}

func (m *pcapManager) fetchBPFConfig() (*pbAgent.BPFConfig, error) {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.fetchBPFConfig")

	logger.Info("getting BPF config")
	m.agentMTLSClientMu.RLock()
	client := m.agentMTLSClient
	m.agentMTLSClientMu.RUnlock()
	if client == nil {
		logger.Error("no mTLS client available, cannot get BPF config")
		return nil, fmt.Errorf("no agent gRPC client available, cannot get BPF config")
	}

	bpfConfig, err := client.GetBPFConfig(m.ctx, &pbAgent.Empty{})
	if err != nil {
		logger.Error("failed to get BPF config", psLog.KeyError, err)
		return nil, err
	}
	return bpfConfig, nil
}

func (m *pcapManager) enforceConfig(bpfConfig *pbAgent.BPFConfig) error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.fetchBPFConfig")

	var errs []error

	if len(bpfConfig.Delete) > 0 {
		for ifaceName, bpfAssociationsToDelete := range bpfConfig.Delete {
			for filterHash, captureCfg := range bpfAssociationsToDelete.Captures {
				deleteErr := m.StopOne(ifaceName, filterHash, captureCfg.Bpf)
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
							slog.String(psLog.KeyBPF, captureCfg.Bpf),
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
			for filterHash, captureCfg := range bpfAssociationsToUpdate.Captures {
				updateErr := m.StopOne(ifaceName, filterHash, captureCfg.Bpf)
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
							slog.String(psLog.KeyBPF, captureCfg.Bpf),
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
						BPF:         captureCfg.Bpf,
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
						slog.String(psLog.KeyBPF, captureCfg.Bpf),
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
			for filterHash, captureCfg := range bpfAssociationsToCreate.Captures {
				createdPacketCapture, createErr := newPacketCapture(
					m.ctx,
					m.logger,
					&CaptureConfig{
						BPF:         captureCfg.Bpf,
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
						slog.String(psLog.KeyBPF, captureCfg.Bpf),
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

func (m *pcapManager) sendPacketEvent(wrappedPkt WrappedPacket) error {
	logger := m.logger.With(psLog.KeyFunction, "PCapManager.sendPacketEvent")

	pkt := wrappedPkt.PacketEventData
	if pkt == nil {
		return fmt.Errorf("no packet event data in wrapped packet")
	}

	if err := pkt.ErrorLayer(); err != nil {
		logger.Error("failed to decode packet", psLog.KeyError, err)
	}

	packetEvent := ConvertPacketToEvent(wrappedPkt)

	m.streamMu.Lock()
	defer m.streamMu.Unlock()

	if m.packetStreamClient == nil {
		logger.Warn("no stream available, dropping packet", psLog.KeyDroppedPacket, pkt.String())
		return nil
	}

	err := m.packetStreamClient.Send(packetEvent)
	if err != nil {
		// TODO: switch on codes for retry-able versus non-retry-able errors
		// create reconnection logic for non-retryable ones
		// send on a reconnectCh to get a new stream client on demand,
		// as opposed to the existing subscriber mechanism
		// for receiving a new client connection when the client cert changes
		logger.Error("failed to send packet over stream", psLog.KeyError, err)
		return err
	}

	return nil
}

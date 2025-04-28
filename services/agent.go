package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
)

type agentService struct {
	pbAgent.UnimplementedAgentServiceServer
	datastore *dao.Datastore
	logger    *slog.Logger
	jetStream nats.JetStream
}

func NewAgentService(js nats.JetStreamContext, datastore *dao.Datastore, logger *slog.Logger) pbAgent.AgentServiceServer {
	return &agentService{
		datastore: datastore,
		jetStream: js,
		logger:    logger,
	}
}

func (as *agentService) ReportInterfaces(ctx context.Context, req *pbAgent.ReportInterfacesRequest) (*pbAgent.Empty, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.ReportInterfaces")

	osUniqueIdentifier, err := as.getSubjectCNFromClientCert(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	existingDevice, err := as.datastore.Devices.GetDeviceByPredicate(postgres.PredicateOSUniqueIdentifier, osUniqueIdentifier)
	if err != nil {
		logger.Error("error looking up device by os_unique_identifier", psLog.KeyError, err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "%s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "%s", fmt.Sprintf("error looking up device by os_unique_identifier: %v", err))
	}
	if existingDevice == nil {
		logger.Error("device record is nil")
		return nil, status.Errorf(codes.Internal, "%s", fmt.Sprintf("device record nil when selected by os_unique_identifier: %s", osUniqueIdentifier))
	}

	interfaces := make([]string, len(req.Interfaces))

	for _, iface := range req.Interfaces {
		logger.Info("received interface name", psLog.KeyDeviceName, iface.Name)
		interfaces = append(interfaces, iface.Name)
	}

	existingDevice.Interfaces = interfaces
	existingDevice.PCapVersion = req.PcapVersion

	err = as.datastore.Devices.Update(existingDevice)
	if err != nil {
		logger.Error("error updating device", psLog.KeyError, err)
		return nil, status.Errorf(codes.Internal, "%s", fmt.Sprintf("error updating device: %v", err))
	}

	return &pbAgent.Empty{}, nil
}

func (as *agentService) PollCommand(ctx context.Context, req *pbAgent.Empty) (*pbAgent.CommandsResponse, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.PollCommand")

	osUniqueIdentifier, err := as.getSubjectCNFromClientCert(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	subject := "cmds." + osUniqueIdentifier
	durable := osUniqueIdentifier
	stream := "COMMANDS"

	// Try to bind to the existing durable consumer
	sub, err := as.jetStream.PullSubscribe(subject, "",
		nats.Bind(stream, durable),
		nats.ManualAck(),
		nats.PullMaxWaiting(128),
	)

	if err == nats.ErrConsumerNotFound {
		// First-time setup: create the durable consumer explicitly
		sub, err = as.jetStream.PullSubscribe(subject, durable,
			nats.BindStream(stream),
			nats.ManualAck(),
			nats.PullMaxWaiting(128),
		)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Pull up to 10 commands or wait 1s
	msgs, err := sub.Fetch(10, nats.MaxWait(1*time.Second))
	if err != nil && err != nats.ErrTimeout {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var commands []string
	for _, msg := range msgs {
		commands = append(commands, string(msg.Data))
		msg.Ack()
	}
	if len(commands) > 0 {
		logger.Info("sending commands received from NATS stream")

		pbCmds := make([]*pbAgent.Command, len(commands))
		for _, cmdStr := range commands {
			pbCmds = append(pbCmds, &pbAgent.Command{
				Name: cmdStr,
			})
		}

		return &pbAgent.CommandsResponse{
			Commands: pbCmds,
		}, nil
	} else {
		logger.Info("no commands on NATS stream, sending noop")
		return &pbAgent.CommandsResponse{
			Commands: []*pbAgent.Command{{
				Name: "noop",
			}},
		}, nil
	}
}

func (as *agentService) GetBPFConfig(ctx context.Context, req *pbAgent.Empty) (*pbAgent.BPFConfig, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.GetBPFConfig")

	osUniqueIdentifier, err := as.getSubjectCNFromClientCert(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	logger.Info("reading device by os_unique_identifier from database")
	device, err := as.datastore.Devices.GetDeviceByPredicate(postgres.PredicateOSUniqueIdentifier, osUniqueIdentifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if device == nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to read device record"))
	}
	return buildBPFConfig(device), nil
}

func (as *agentService) SendPacketEvent(stream pbAgent.AgentService_SendPacketEventServer) error {
	logger := as.logger.With(psLog.KeyFunction, "agentService.SendPacketEvent")

	for {
		packet, err := stream.Recv()
		if err != nil {
			if err == context.Canceled || status.Code(err) == codes.Canceled {
				logger.Info("stream context canceled (likely client disconnect)")
				return nil
			}
			if err == io.EOF {
				logger.Info("received EOF from packet stream")
				return stream.SendAndClose(&pbAgent.Empty{})
			}
			logger.Error("error receiving from packet stream", "error", err)
			return err
		}

		layers := packet.Layers
		if layers != nil {
			if layers.IpLayer != nil {
				ipLayer := layers.IpLayer
				logger.Info(fmt.Sprintf("IP src %s for dst %s", ipLayer.SrcIp, ipLayer.DstIp))
			}
			if layers.TcpLayer != nil {
				tcpLayer := layers.TcpLayer
				logger.Info(fmt.Sprintf("TCP src port %d for dst port %d", tcpLayer.SrcPort, tcpLayer.DstPort))
			}
			if layers.TlsLayer != nil {
				tlsLayer := layers.TlsLayer
				for _, record := range tlsLayer.Records {
					logger.Info(fmt.Sprintf("tls record type %s", record.Type))
				}
			}
			if layers.UdpLayer != nil {
				udpLayer := layers.UdpLayer
				logger.Info(fmt.Sprintf("upd src port %d for dst port %d", udpLayer.SrcPort, udpLayer.DstPort))
			}
		}
	}
}

func (as *agentService) getSubjectCNFromClientCert(ctx context.Context) (string, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.getSubjectCNFromClientCert")
	logger.Info("getting peer from context")
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no peer info")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return "", fmt.Errorf("no client certificate")
	}

	clientCert := tlsInfo.State.PeerCertificates[0]
	return clientCert.Subject.CommonName, nil
}

func buildBPFConfig(device *dao.Device) *pbAgent.BPFConfig {
	create := make(map[string]*pbAgent.InterfaceCaptureMap)
	update := make(map[string]*pbAgent.InterfaceCaptureMap)
	delete := make(map[string]*pbAgent.InterfaceCaptureMap)

	// Helper to deep-copy a dao.CaptureConfig to pbAgent.CaptureConfig
	convertConfig := func(c dao.CaptureConfig) *pbAgent.CaptureConfig {
		return &pbAgent.CaptureConfig{
			Bpf:         c.Bpf,
			DeviceName:  c.DeviceName,
			Promiscuous: c.Promiscuous,
			SnapLen:     c.SnapLen,
		}
	}

	// Helper to add a capture into a map
	addCapture := func(m map[string]*pbAgent.InterfaceCaptureMap, iface string, bpfID uint64, config dao.CaptureConfig) {
		_, exists := m[iface]
		if !exists {
			m[iface] = &pbAgent.InterfaceCaptureMap{Captures: make(map[uint64]*pbAgent.CaptureConfig)}
		}
		m[iface].Captures[bpfID] = convertConfig(config)
	}

	// Build sets for faster lookup
	current := device.InterfaceBPFAssociations
	previous := device.PreviousAssociations

	// First pass: detect Creates and Updates
	for currentIface, currentCaptures := range current {
		prevCaptures, currentIfaceExistsInPrevious := previous[currentIface]
		if !currentIfaceExistsInPrevious {
			// the interface name did not have an entry in the previous associations
			// so no need to check the nested map,
			// just add all capture configs for this interface to the create map
			for currentBPFHash, currentConfig := range currentCaptures {
				addCapture(create, currentIface, currentBPFHash, currentConfig)
			}
			continue
		}

		for currentBPFHash, currentConfig := range currentCaptures {
			prevConfig, currentBPFHashExistsInPrevious := prevCaptures[currentBPFHash]
			if !currentBPFHashExistsInPrevious {
				// the interface name did have an entry in the outer map, but the BPF hash did not,
				// so add it to the create map
				addCapture(create, currentIface, currentBPFHash, currentConfig)
			} else {
				// the interface name and the BPF hash existed in outer and inner maps, respectively,
				// so check for delta and, if there is a diff, add to the update map
				if configsDiffer(prevConfig, currentConfig) {
					addCapture(update, currentIface, currentBPFHash, currentConfig)
				}
			}
		}
	}

	// Second pass: detect Deletes
	for previousIface, prevCaptures := range previous {
		currCaptures, previousIfaceExistsInCurrent := current[previousIface]
		if !previousIfaceExistsInCurrent {
			// the interface name did have an entry in previous and has no entry in current
			// no need to check the existence of the BPF hashes in the nested map
			// we have to delete all BPF hashes for this interface
			for bpfID := range prevCaptures {
				addCapture(delete, previousIface, bpfID, prevCaptures[bpfID])
			}
			continue
		}

		for prevBPFHash := range prevCaptures {
			_, previousBPFHashExistsInCurrent := currCaptures[prevBPFHash]
			if !previousBPFHashExistsInCurrent {
				// the interface name has an entry in both previous and current
				// but the BPF hash from the previous nested map has no corresponding entry
				// in the current nested map, so add it to the delete map
				addCapture(delete, previousIface, prevBPFHash, prevCaptures[prevBPFHash])
			}
		}
	}

	return &pbAgent.BPFConfig{
		Create: create,
		Update: update,
		Delete: delete,
	}
}

// configsDiffer compares two dao.CaptureConfig structs for any differences
func configsDiffer(a, b dao.CaptureConfig) bool {
	return a.Bpf != b.Bpf ||
		a.DeviceName != b.DeviceName ||
		a.Promiscuous != b.Promiscuous ||
		a.SnapLen != b.SnapLen
}

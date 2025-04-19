package agent

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/cespare/xxhash"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
)

type agentService struct {
	pbAgent.UnimplementedAgentServiceServer
	// TODO: obviously this is just for PoC purposes
	commandCounter   int
	bpfCounter       int
	logger           *slog.Logger
	testBPFResponses []*pbAgent.BPFConfig
}

func buildInterfaceCaptureMap(captures map[uint64]*pbAgent.CaptureConfig) *pbAgent.InterfaceCaptureMap {
	return &pbAgent.InterfaceCaptureMap{
		Captures: captures,
	}
}

func NewAgentService(logger *slog.Logger) pbAgent.AgentServiceServer {
	/*
		linux loopback is lo
		darwin loopback is lo0
		windows loopback is \Device\NPF_Loopback
	*/
	testBPFResponses := []*pbAgent.BPFConfig{
		{
			Create: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("tcp port 3000"): {
						Bpf:         "tcp port 3000",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
					xxhash.Sum64String("udp port 53"): {
						Bpf:         "udp port 53",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
				}),
			},
			Update: map[string]*pbAgent.InterfaceCaptureMap{},
			Delete: map[string]*pbAgent.InterfaceCaptureMap{},
		},
		{
			Create: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("icmp"): {
						Bpf:         "icmp",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
				}),
			},
			Update: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("tcp port 3000"): {
						Bpf:         "tcp port 3000",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
						Timeout:     int64(time.Second * 2),
					},
				}),
			},
			Delete: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("upd port 53"): {
						Bpf:         "upd port 53",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
				}),
			},
		},
		{
			Create: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("udp port 5353"): {
						Bpf:         "udp port 5353",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
				}),
			},
			Update: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("icmp"): {
						Bpf:         "icmp",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     8192,
					},
				}),
			},
			Delete: map[string]*pbAgent.InterfaceCaptureMap{
				"lo": buildInterfaceCaptureMap(map[uint64]*pbAgent.CaptureConfig{
					xxhash.Sum64String("tcp port 3000"): {
						Bpf:         "tcp port 3000",
						DeviceName:  "lo",
						Promiscuous: false,
						SnapLen:     65535,
					},
				}),
			},
		},
	}

	return &agentService{
		logger:           logger,
		testBPFResponses: testBPFResponses,
	}
}

func (as *agentService) ReportInterfaces(ctx context.Context, req *pbAgent.ReportInterfacesRequest) (*pbAgent.Empty, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.ReportInterfaces")

	for _, iface := range req.Interfaces {
		logger.Info("received interface name", psLog.KeyDeviceName, iface.Name)
	}
	return &pbAgent.Empty{}, nil
}

func (as *agentService) PollCommand(ctx context.Context, req *pbAgent.Empty) (*pbAgent.Command, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.PollCommand")

	// TODO: this is for testing purposes
	logger.Info("received poll request, sending command", "commandCounter", as.commandCounter)
	if as.commandCounter == 0 {
		as.commandCounter += 1
		return &pbAgent.Command{
			Name: "send_interfaces",
		}, nil
	}
	return &pbAgent.Command{
		Name: "get_bpf_config",
	}, nil
}
func (as *agentService) GetBPFConfig(ctx context.Context, req *pbAgent.Empty) (*pbAgent.BPFConfig, error) {
	logger := as.logger.With(psLog.KeyFunction, "agentService.GetBPFConfig")

	resp := as.testBPFResponses[as.bpfCounter]
	logger.Info("sending BPF response", "bpfResponseCounter", as.bpfCounter)
	fmt.Printf("sending BPF response index %d\n", as.bpfCounter)
	as.bpfCounter = (as.bpfCounter + 1) % len(as.testBPFResponses)
	return resp, nil
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

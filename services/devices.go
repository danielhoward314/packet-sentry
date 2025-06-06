package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres"
	pbDevices "github.com/danielhoward314/packet-sentry/protogen/golang/devices"
	"github.com/nats-io/nats.go"
)

const (
	svcNameDevices = "devices"
)

// devicesService implements the devices gRPC service
type devicesService struct {
	pbDevices.UnimplementedDevicesServiceServer
	datastore *dao.Datastore
	jetStream nats.JetStream
	logger    *slog.Logger
}

func NewDevicesService(
	datastore *dao.Datastore,
	js nats.JetStreamContext,
	baseLogger *slog.Logger,
) pbDevices.DevicesServiceServer {
	childLogger := baseLogger.With(slog.String("service", svcNameDevices))

	return &devicesService{
		datastore: datastore,
		jetStream: js,
		logger:    childLogger,
	}
}

func (ds *devicesService) Get(ctx context.Context, request *pbDevices.GetDeviceRequest) (*pbDevices.GetDeviceResponse, error) {
	if request.Id == "" {
		ds.logger.Error("invalid device id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid device id")
	}
	device, err := ds.datastore.Devices.GetDeviceByPredicate(postgres.PredicateID, request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "device not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read device data: %s", err.Error())
	}
	if device == nil {
		return nil, status.Error(codes.Internal, "failed to read device data")
	}

	pbAssociations := make(map[string]*pbDevices.InterfaceCaptureMap)
	pbPreviousAssociations := make(map[string]*pbDevices.InterfaceCaptureMap)

	for ifaceName, daoInterfaceToBPFMap := range device.InterfaceBPFAssociations {
		_, ok := pbAssociations[ifaceName]
		if !ok {
			pbAssociations[ifaceName] = &pbDevices.InterfaceCaptureMap{
				Captures: make(map[uint64]*pbDevices.CaptureConfig),
			}
		}
		for daoBPFHash, daoCaptureConfig := range daoInterfaceToBPFMap {
			if pbAssociations[ifaceName].Captures == nil {
				pbAssociations[ifaceName].Captures = make(map[uint64]*pbDevices.CaptureConfig)
			}
			pbAssociations[ifaceName].Captures[daoBPFHash] = &pbDevices.CaptureConfig{
				Bpf:         daoCaptureConfig.Bpf,
				DeviceName:  daoCaptureConfig.DeviceName,
				Promiscuous: daoCaptureConfig.Promiscuous,
				SnapLen:     int32(daoCaptureConfig.SnapLen),
			}
		}
	}
	for ifaceName, daoPreviousAssociations := range device.PreviousAssociations {
		_, ok := pbPreviousAssociations[ifaceName]
		if !ok {
			pbPreviousAssociations[ifaceName] = &pbDevices.InterfaceCaptureMap{
				Captures: make(map[uint64]*pbDevices.CaptureConfig),
			}
		}
		for daoPreviousBPFHash, daoPreviousCaptureConfig := range daoPreviousAssociations {
			if pbPreviousAssociations[ifaceName].Captures == nil {
				pbPreviousAssociations[ifaceName].Captures = make(map[uint64]*pbDevices.CaptureConfig)
			}
			pbPreviousAssociations[ifaceName].Captures[daoPreviousBPFHash] = &pbDevices.CaptureConfig{
				Bpf:         daoPreviousCaptureConfig.Bpf,
				DeviceName:  daoPreviousCaptureConfig.DeviceName,
				Promiscuous: daoPreviousCaptureConfig.Promiscuous,
				SnapLen:     int32(daoPreviousCaptureConfig.SnapLen),
			}
		}
	}
	return &pbDevices.GetDeviceResponse{
		Id:                       device.ID,
		OrganizationId:           device.OrganizationID,
		OsUniqueIdentifier:       device.OSUniqueIdentifier,
		ClientCertPem:            device.ClientCertPEM,
		ClientCertFingerprint:    device.ClientCertFingerprint,
		InterfaceBpfAssociations: pbAssociations,
		PreviousAssociations:     pbPreviousAssociations,
		PcapVersion:              device.PCapVersion,
		Interfaces:               device.Interfaces,
	}, nil
}

func (ds *devicesService) List(ctx context.Context, request *pbDevices.ListDevicesRequest) (*pbDevices.ListDevicesResponse, error) {
	if request.OrganizationId == "" {
		ds.logger.Error("invalid organization id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization id")
	}

	devices, err := ds.datastore.Devices.List(request.OrganizationId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "devices not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read devices data: %s", err.Error())
	}

	response := &pbDevices.ListDevicesResponse{
		Devices: make([]*pbDevices.GetDeviceResponse, 0, len(devices)),
	}

	for _, device := range devices {
		pbAssociations := make(map[string]*pbDevices.InterfaceCaptureMap)
		pbPreviousAssociations := make(map[string]*pbDevices.InterfaceCaptureMap)

		for ifaceName, daoInterfaceToBPFMap := range device.InterfaceBPFAssociations {
			_, ok := pbAssociations[ifaceName]
			if !ok {
				pbAssociations[ifaceName] = &pbDevices.InterfaceCaptureMap{
					Captures: make(map[uint64]*pbDevices.CaptureConfig),
				}
			}
			for daoBPFHash, daoCaptureConfig := range daoInterfaceToBPFMap {
				if pbAssociations[ifaceName].Captures == nil {
					pbAssociations[ifaceName].Captures = make(map[uint64]*pbDevices.CaptureConfig)
				}
				pbAssociations[ifaceName].Captures[daoBPFHash] = &pbDevices.CaptureConfig{
					Bpf:         daoCaptureConfig.Bpf,
					DeviceName:  daoCaptureConfig.DeviceName,
					Promiscuous: daoCaptureConfig.Promiscuous,
					SnapLen:     int32(daoCaptureConfig.SnapLen),
				}
			}
		}
		for ifaceName, daoPreviousAssociations := range device.PreviousAssociations {
			_, ok := pbPreviousAssociations[ifaceName]
			if !ok {
				pbPreviousAssociations[ifaceName] = &pbDevices.InterfaceCaptureMap{
					Captures: make(map[uint64]*pbDevices.CaptureConfig),
				}
			}
			for daoPreviousBPFHash, daoPreviousCaptureConfig := range daoPreviousAssociations {
				if pbPreviousAssociations[ifaceName].Captures == nil {
					pbPreviousAssociations[ifaceName].Captures = make(map[uint64]*pbDevices.CaptureConfig)
				}
				pbPreviousAssociations[ifaceName].Captures[daoPreviousBPFHash] = &pbDevices.CaptureConfig{
					Bpf:         daoPreviousCaptureConfig.Bpf,
					DeviceName:  daoPreviousCaptureConfig.DeviceName,
					Promiscuous: daoPreviousCaptureConfig.Promiscuous,
					SnapLen:     int32(daoPreviousCaptureConfig.SnapLen),
				}
			}
		}
		response.Devices = append(response.Devices, &pbDevices.GetDeviceResponse{
			Id:                       device.ID,
			OrganizationId:           device.OrganizationID,
			OsUniqueIdentifier:       device.OSUniqueIdentifier,
			ClientCertPem:            device.ClientCertPEM,
			ClientCertFingerprint:    device.ClientCertFingerprint,
			InterfaceBpfAssociations: pbAssociations,
			PreviousAssociations:     pbPreviousAssociations,
			PcapVersion:              device.PCapVersion,
			Interfaces:               device.Interfaces,
		})
	}

	return response, nil
}

// Update expects all fields provided
func (ds *devicesService) Update(ctx context.Context, request *pbDevices.UpdateDeviceRequest) (*pbDevices.Empty, error) {
	if request.Id == "" {
		ds.logger.Error("invalid device id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid device id")
	}
	device, err := ds.datastore.Devices.GetDeviceByPredicate(postgres.PredicateID, request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "device not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read device data: %s", err.Error())
	}
	if device == nil {
		return nil, status.Error(codes.Internal, "failed to read device data")
	}
	if request.ClientCertPem != "" {
		device.ClientCertPEM = request.ClientCertPem
	}
	if request.ClientCertFingerprint != "" {
		device.ClientCertFingerprint = request.ClientCertFingerprint
	}
	if request.ClientCertFingerprint != "" {
		device.ClientCertFingerprint = request.ClientCertFingerprint
	}
	if len(request.Interfaces) > 0 {
		device.Interfaces = request.Interfaces
	}
	if request.PcapVersion != "" {
		device.PCapVersion = request.PcapVersion
	}

	device.PreviousAssociations = device.InterfaceBPFAssociations

	daoAssociations := make(map[string]map[uint64]dao.CaptureConfig)

	for ifaceName, pbInterfaceToBPFMap := range request.InterfaceBpfAssociations {
		_, ok := request.InterfaceBpfAssociations[ifaceName]
		if !ok {
			daoAssociations[ifaceName] = make(map[uint64]dao.CaptureConfig)
		}
		for pbBPF, pbCaptureConfig := range pbInterfaceToBPFMap.Captures {
			if daoAssociations[ifaceName] == nil {
				daoAssociations[ifaceName] = make(map[uint64]dao.CaptureConfig)
			}
			pbBPFHash := xxhash.Sum64([]byte(pbBPF))
			daoAssociations[ifaceName][pbBPFHash] = dao.CaptureConfig{
				Bpf:         pbCaptureConfig.Bpf,
				DeviceName:  pbCaptureConfig.DeviceName,
				Promiscuous: pbCaptureConfig.Promiscuous,
				SnapLen:     int32(65535),
			}
		}
	}

	device.InterfaceBPFAssociations = daoAssociations

	err = ds.datastore.Devices.Update(device)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	_, err = ds.jetStream.Publish("cmds."+device.OSUniqueIdentifier, []byte("get_bpf_config"))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", fmt.Sprintf("command send was not ack'd: %v", err))
	}
	return &pbDevices.Empty{}, nil
}

package services

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/danielhoward314/packet-sentry/dao"
	pbEvents "github.com/danielhoward314/packet-sentry/protogen/golang/events"
)

const (
	svcNameEvents = "events"
)

// eventsService implements the events gRPC service
type eventsService struct {
	pbEvents.UnimplementedEventsServiceServer
	datastore *dao.TimescaleDatastore
	logger    *slog.Logger
}

func NewEventsService(
	datastore *dao.TimescaleDatastore,
	baseLogger *slog.Logger,
) pbEvents.EventsServiceServer {
	childLogger := baseLogger.With(slog.String("service", svcNameEvents))

	return &eventsService{
		datastore: datastore,
		logger:    childLogger,
	}
}

func (es *eventsService) Get(ctx context.Context, request *pbEvents.GetEventsRequest) (*pbEvents.GetEventsResponse, error) {
	if request.DeviceId == "" {
		es.logger.Error("invalid device id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid device id")
	}
	if request.End == "" {
		es.logger.Error("invalid end datetime query string")
		return nil, status.Errorf(codes.InvalidArgument, "invalid end datetime query string")
	}
	if request.Start == "" {
		es.logger.Error("invalid start datetime query string")
		return nil, status.Errorf(codes.InvalidArgument, "invalid start datetime query string")
	}

	es.logger.Info("querying events", "os_unique_identifier", request.DeviceId, "start", request.Start, "end", request.End)
	events, err := es.datastore.Events.Read(
		request.DeviceId,
		request.Start,
		request.End,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "events not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read events data: %s", err.Error())
	}
	if events == nil {
		return nil, status.Error(codes.Internal, "failed to read events data")
	}

	resEvents := make([]*pbEvents.Event, 0)
	for _, event := range events {
		resEvents = append(resEvents, &pbEvents.Event{
			EventTime:      event.EventTime,
			Bpf:            event.Bpf,
			OriginalLength: event.OriginalLength,
			IpSrc:          event.IpSrc,
			IpDst:          event.IpDst,
			TcpSrcPort:     event.TcpSrcPort,
			TcpDstPort:     event.TcpDstPort,
			IpVersion:      event.IpVersion,
		})
	}

	return &pbEvents.GetEventsResponse{
		Events: resEvents,
	}, nil
}

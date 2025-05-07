package services

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/danielhoward314/packet-sentry/dao"
	pbOrgs "github.com/danielhoward314/packet-sentry/protogen/golang/organizations"
)

const (
	svcNameOrganizations = "organizations"
)

// organizationsService implements the organizations gRPC service
type organizationsService struct {
	pbOrgs.UnimplementedOrganizationsServiceServer
	datastore *dao.Datastore
	logger    *slog.Logger
}

func NewOrganizationsService(
	datastore *dao.Datastore,
	baseLogger *slog.Logger,
) pbOrgs.OrganizationsServiceServer {
	childLogger := baseLogger.With(slog.String("service", svcNameOrganizations))

	return &organizationsService{
		datastore: datastore,
		logger:    childLogger,
	}
}

func (os *organizationsService) Get(ctx context.Context, request *pbOrgs.GetOrganizationRequest) (*pbOrgs.GetOrganizationResponse, error) {
	if request.Id == "" {
		os.logger.Error("invalid organization id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization id")
	}
	org, err := os.datastore.Organizations.Read(request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "organization not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read organization data: %s", err.Error())
	}
	return &pbOrgs.GetOrganizationResponse{
		Id:               org.ID,
		OrganizationName: org.Name,
		BillingPlan:      org.BillingPlanType,
	}, nil
}

package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/danielhoward314/packet-sentry/dao"
	orgspb "github.com/danielhoward314/packet-sentry/protogen/golang/organizations"
)

// organizationsService implements the organizations gRPC service
type organizationsService struct {
	orgspb.UnimplementedOrganizationsServiceServer
	datastore *dao.Datastore
}

func NewOrganizationsService(
	datastore *dao.Datastore,
) orgspb.OrganizationsServiceServer {
	return &organizationsService{
		datastore: datastore,
	}
}

func (os *organizationsService) Get(ctx context.Context, request *orgspb.GetOrganizationRequest) (*orgspb.GetOrganizationResponse, error) {
	if request.Id == "" {
		slog.Error("invalid organization id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization id")
	}
	org, err := os.datastore.Organizations.Read(request.Id)
	fmt.Printf("request.Id: %s\n", request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "organization not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read organization data: %s", err.Error())
	}
	return &orgspb.GetOrganizationResponse{
		Id:               org.ID,
		OrganizationName: org.Name,
		BillingPlan:      org.BillingPlanType,
	}, nil
}

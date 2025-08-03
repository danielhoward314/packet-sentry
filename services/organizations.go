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

	maskedCreditCard := ""
	if org.PaymentDetails != nil && len(org.PaymentDetails.CardNumber) >= 4 {
		maskedCreditCard = org.PaymentDetails.CardNumber[len(org.PaymentDetails.CardNumber)-4:]

	}

	return &pbOrgs.GetOrganizationResponse{
		Id:                        org.ID,
		OrganizationName:          org.Name,
		BillingPlan:               org.BillingPlanType,
		PrimaryAdministratorEmail: org.PrimaryAdministratorEmail,
		MaskedCreditCard:          maskedCreditCard,
	}, nil
}

func (os *organizationsService) Update(ctx context.Context, request *pbOrgs.UpdateOrganizationRequest) (*pbOrgs.Empty, error) {
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

	if request.Name != "" {
		org.Name = request.Name
	}
	if request.BillingPlan != "" {
		org.BillingPlanType = request.BillingPlan
	}

	if request.PaymentDetails != nil {
		if request.PaymentDetails.CardName == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.AddressLineOne == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.AddressLineTwo == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.CardNumber == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.ExpirationMonth == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.ExpirationYear == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		if request.PaymentDetails.Cvc == "" {
			os.logger.Error("invalid payment details")
			return nil, status.Errorf(codes.InvalidArgument, "invalid payment details")
		}
		org.PaymentDetails = &dao.PaymentDetails{
			CardName:        request.PaymentDetails.CardName,
			AddressLineOne:  request.PaymentDetails.AddressLineOne,
			AddressLineTwo:  request.PaymentDetails.AddressLineTwo,
			CardNumber:      request.PaymentDetails.CardNumber,
			ExpirationMonth: request.PaymentDetails.ExpirationMonth,
			ExpirationYear:  request.PaymentDetails.ExpirationYear,
			Cvc:             request.PaymentDetails.Cvc,
		}
	}

	err = os.datastore.Organizations.Update(org)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update organization data: %s", err.Error())
	}
	return &pbOrgs.Empty{}, nil
}

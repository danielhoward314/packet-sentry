package services

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/hashes"
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
	pbAuth "github.com/danielhoward314/packet-sentry/protogen/golang/auth"
)

// authService implements the account gRPC service
type authService struct {
	pbAuth.UnimplementedAuthServiceServer
	datastore      *dao.Datastore
	tokenDatastore dao.TokenDatastore
	smtpDialer     *gomail.Dialer
}

func NewAuthService(
	datastore *dao.Datastore,
	tokenDatastore dao.TokenDatastore,
	smtpDialer *gomail.Dialer,
) pbAuth.AuthServiceServer {
	return &authService{
		datastore:      datastore,
		tokenDatastore: tokenDatastore,
		smtpDialer:     smtpDialer,
	}
}

// ValidateSession validates admin ui session data submitted via a JWT in the request
func (as *authService) ValidateSession(ctx context.Context, request *pbAuth.ValidateSessionRequest) (*pbAuth.ValidateSessionResponse, error) {
	if request.Jwt == "" {
		slog.Error("invalid session JWT")
		return nil, status.Errorf(codes.InvalidArgument, "invalid session JWT")
	}
	sessionTokenData, err := as.tokenDatastore.Read(request.Jwt)
	if err != nil {
		if err == redis.Nil {
			return nil, status.Errorf(codes.NotFound, "session data not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read session data: %s", err.Error())
	}
	administrator, err := as.datastore.Administrators.Read(sessionTokenData.AdministratorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	if !administrator.Verified {
		return nil, status.Errorf(codes.PermissionDenied, "email not verified")
	}
	err = as.tokenDatastore.Decode(psJWT.Access, request.Jwt, psJWT.AdminUISession)
	if err != nil {
		if err.Error() == psJWT.TokenExpiredError {
			return nil, status.Errorf(codes.Unauthenticated, "access token has expired, use refresh token to request another")
		}
		if err.Error() == psJWT.InvalidTokenError {
			return nil, status.Errorf(codes.PermissionDenied, "invalid access token")
		}
		return nil, status.Errorf(codes.Internal, "failed to validate session JWT: %s", err.Error())
	}
	return &pbAuth.ValidateSessionResponse{
		Jwt: request.Jwt,
	}, nil
}

func (as *authService) Login(ctx context.Context, request *pbAuth.LoginRequest) (*pbAuth.LoginResponse, error) {
	if request.Email == "" {
		slog.Error("invalid email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if request.Password == "" {
		slog.Error("invalid password")
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}
	administrator, err := as.datastore.Administrators.ReadByEmail(request.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	if !administrator.Verified {
		return nil, status.Errorf(codes.PermissionDenied, "email not verified")
	}
	err = hashes.ValidateBCryptHashedCleartext(administrator.PasswordHash, request.Password)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "authentication error")
	}
	organization, err := as.datastore.Organizations.Read(administrator.OrganizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "organization not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read organization data: %s", err.Error())
	}
	adminUIAccessToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			AdministratorID:   administrator.ID,
			OrganizationID:    administrator.OrganizationID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Access,
			ClaimsType:        psJWT.AdminUISession,
		},
		psJWT.Access,
		psJWT.AdminUISession,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err.Error())
	}
	adminUIRefreshToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			AdministratorID:   administrator.ID,
			OrganizationID:    administrator.OrganizationID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Refresh,
			ClaimsType:        psJWT.AdminUISession,
		},
		psJWT.Refresh,
		psJWT.AdminUISession,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err.Error())
	}
	apiAccessToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			AdministratorID:   administrator.ID,
			OrganizationID:    administrator.OrganizationID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Access,
			ClaimsType:        psJWT.APIAuthorization,
		},
		psJWT.Access,
		psJWT.APIAuthorization,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err.Error())
	}
	apiRefreshToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			AdministratorID:   administrator.ID,
			OrganizationID:    administrator.OrganizationID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Refresh,
			ClaimsType:        psJWT.APIAuthorization,
		},
		psJWT.Refresh,
		psJWT.APIAuthorization,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err.Error())
	}
	return &pbAuth.LoginResponse{
		AdministratorId:     administrator.ID,
		OrganizationId:      administrator.OrganizationID,
		AdministratorName:   administrator.DisplayName,
		OrganizationName:    organization.Name,
		BillingPlan:         organization.BillingPlanType,
		AdminUiAccessToken:  adminUIAccessToken,
		AdminUiRefreshToken: adminUIRefreshToken,
		ApiAccessToken:      apiAccessToken,
		ApiRefreshToken:     apiRefreshToken,
	}, nil
}

// RefreshToken takes in a refesh JWT of a given claims type and, if valid, returns a new access JWT of the same claims type
func (as *authService) RefreshToken(ctx context.Context, request *pbAuth.RefreshTokenRequest) (*pbAuth.RefreshTokenResponse, error) {
	if request.Jwt == "" {
		slog.Error("invalid refresh JWT")
		return nil, status.Errorf(codes.InvalidArgument, "invalid refresh JWT")
	}
	claimsType, err := psJWT.GetClaimsTypeFromProtoEnum(request.ClaimsType)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %s", err.Error())
	}
	refreshTokenData, err := as.tokenDatastore.Read(request.Jwt)
	if err != nil {
		if err == redis.Nil {
			return nil, status.Errorf(codes.NotFound, "refresh token data not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read refresh token data: %s", err.Error())
	}
	administrator, err := as.datastore.Administrators.Read(refreshTokenData.AdministratorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	if !administrator.Verified {
		return nil, status.Errorf(codes.PermissionDenied, "email not verified")
	}
	err = as.tokenDatastore.Decode(psJWT.Refresh, request.Jwt, claimsType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate refresh JWT: %s", err.Error())
	}
	// use the same tokenData as what was used for refresh token
	// tokenType hard-coded to access, since a refresh is always used to get an access token
	accessJWT, err := as.tokenDatastore.Create(refreshTokenData, psJWT.Access, claimsType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access JWT: %s", err.Error())
	}
	return &pbAuth.RefreshTokenResponse{Jwt: accessJWT}, nil
}

func (as *authService) CreateInstallKey(ctx context.Context, request *pbAuth.CreateInstallKeyRequest) (*pbAuth.CreateInstallKeyResponse, error) {
	if request.AdministratorEmail == "" {
		slog.Error("invalid administrator email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid administrator email")
	}

	administrator, err := as.datastore.Administrators.ReadByEmail(request.AdministratorEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	if !administrator.Verified {
		return nil, status.Errorf(codes.PermissionDenied, "administrator email not verified")
	}
	installKey, err := as.datastore.InstallKeys.Create(administrator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create install key: %s", err.Error())
	}
	return &pbAuth.CreateInstallKeyResponse{
		InstallKey: installKey,
	}, nil
}

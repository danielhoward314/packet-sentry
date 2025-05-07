package services

import (
	"bytes"
	"context"
	"database/sql"
	"html/template"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres"
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
	pbAccounts "github.com/danielhoward314/packet-sentry/protogen/golang/accounts"
)

const (
	emailFrom       = "no.reply@packet-sentry.com"
	svcNameAccounts = "accounts"
)

// accountsService implements the account gRPC service
type accountsService struct {
	pbAccounts.UnimplementedAccountsServiceServer
	datastore             *dao.Datastore
	logger                *slog.Logger
	registrationDatastore dao.RegistrationDatastore
	tokenDatastore        dao.TokenDatastore
	smtpDialer            *gomail.Dialer
}

func NewAccountsService(
	datastore *dao.Datastore,
	baseLogger *slog.Logger,
	registrationDatastore dao.RegistrationDatastore,
	tokenDatastore dao.TokenDatastore,
	smtpDialer *gomail.Dialer,
) pbAccounts.AccountsServiceServer {
	childLogger := baseLogger.With(slog.String("service", svcNameAccounts))

	return &accountsService{
		datastore:             datastore,
		logger:                childLogger,
		registrationDatastore: registrationDatastore,
		tokenDatastore:        tokenDatastore,
		smtpDialer:            smtpDialer,
	}
}

// Signup creates a new organization and admin, and triggers primary admin email verification
func (as *accountsService) Signup(ctx context.Context, request *pbAccounts.SignupRequest) (*pbAccounts.SignupResponse, error) {
	if request.OrganizationName == "" {
		as.logger.Error("invalid organization name")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization name")
	}
	if request.PrimaryAdministratorEmail == "" {
		as.logger.Error("invalid primary administrator email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid primary administrator email")
	}
	if request.PrimaryAdministratorName == "" {
		as.logger.Error("invalid primary administrator name")
		return nil, status.Errorf(codes.InvalidArgument, "invalid primary administrator name")
	}
	if request.PrimaryAdministratorCleartextPassword == "" {
		as.logger.Error("invalid primary administrator cleartext password")
		return nil, status.Errorf(codes.InvalidArgument, "invalid primary administrator cleartext password")
	}
	organization := &dao.Organization{
		Name:                      request.OrganizationName,
		PrimaryAdministratorEmail: request.PrimaryAdministratorEmail,
	}
	as.logger.Info("creating organization")
	organizationID, err := as.datastore.Organizations.Create(organization)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create organization")
	}
	administrator := &dao.Administrator{
		Email:             request.PrimaryAdministratorEmail,
		DisplayName:       request.PrimaryAdministratorName,
		OrganizationID:    organizationID,
		AuthorizationRole: postgres.PrimaryAdmin,
	}
	as.logger.Info("creating primary administrator", "organization_id", organizationID)
	administratorID, err := as.datastore.Administrators.Create(administrator, request.PrimaryAdministratorCleartextPassword)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create administrator")
	}
	as.logger.Info("creating registration data", "organization_id", organizationID, "administrator_id", administratorID)
	token, emailCode, err := as.registrationDatastore.Create(&dao.Registration{
		OrganizationID:  organizationID,
		AdministratorID: administratorID,
	})
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create registration")
	}
	emailTemplateData := struct {
		Code string
	}{
		Code: emailCode,
	}
	as.logger.Info("parsing verification email template")
	tmpl, err := template.ParseFiles("templates/verify_email.html")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse registration email template %s", err.Error())
	}
	var body bytes.Buffer
	as.logger.Info("executing verification email template")
	err = tmpl.Execute(&body, emailTemplateData)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to executing registration email template %s", err.Error())
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", request.PrimaryAdministratorEmail)
	m.SetHeader("Subject", "Packet Sentry: Verify your email")
	m.SetBody("text/html", body.String())
	as.logger.Info("sending verification email")
	err = as.smtpDialer.DialAndSend(m)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to send administrator email verification email %s", err.Error())
	}
	return &pbAccounts.SignupResponse{
		Token: token,
	}, nil
}

// Verify validates email verification codes, updates the administrators.verified column & creates admin UI & API JWTs
func (as *accountsService) Verify(ctx context.Context, request *pbAccounts.VerificationRequest) (*pbAccounts.VerificationResponse, error) {
	if request.Token == "" {
		as.logger.Error("invalid token")
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}
	if request.VerificationCode == "" {
		as.logger.Error("invalid verification code")
		return nil, status.Errorf(codes.InvalidArgument, "invalid verification code")
	}
	registration, err := as.registrationDatastore.Read(request.Token)
	if err != nil {
		if err == redis.Nil {
			return nil, status.Errorf(codes.NotFound, "registration token not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read registration token: %s", err.Error())
	}
	if registration.EmailCode != request.VerificationCode {
		return nil, status.Errorf(codes.PermissionDenied, "verification code not authorized")
	}
	administrator, err := as.datastore.Administrators.Read(registration.AdministratorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	administrator.Verified = true
	err = as.datastore.Administrators.Update(administrator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update administrator data: %s", err.Error())
	}
	err = as.registrationDatastore.Delete(request.Token)
	if err != nil {
		// non-fatal error, the registration data has a short TTL
		as.logger.Warn("failed to delete registration data")
	}
	adminUIAccessToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			OrganizationID:    administrator.OrganizationID,
			AdministratorID:   administrator.ID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Access,
			ClaimsType:        psJWT.AdminUISession,
		},
		psJWT.Access,
		psJWT.AdminUISession,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate jwt: %s", err.Error())
	}
	adminUIRefreshToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			OrganizationID:    administrator.OrganizationID,
			AdministratorID:   administrator.ID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Refresh,
			ClaimsType:        psJWT.AdminUISession,
		},
		psJWT.Refresh,
		psJWT.AdminUISession,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate jwt: %s", err.Error())
	}
	apiAccessToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			OrganizationID:    administrator.OrganizationID,
			AdministratorID:   administrator.ID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Access,
			ClaimsType:        psJWT.APIAuthorization,
		},
		psJWT.Access,
		psJWT.APIAuthorization,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate jwt: %s", err.Error())
	}
	apiRefreshToken, err := as.tokenDatastore.Create(
		&dao.TokenData{
			OrganizationID:    administrator.OrganizationID,
			AdministratorID:   administrator.ID,
			AuthorizationRole: administrator.AuthorizationRole,
			TokenType:         psJWT.Refresh,
			ClaimsType:        psJWT.APIAuthorization,
		},
		psJWT.Refresh,
		psJWT.APIAuthorization,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate jwt: %s", err.Error())
	}
	return &pbAccounts.VerificationResponse{
		AdminUiAccessToken:  adminUIAccessToken,
		AdminUiRefreshToken: adminUIRefreshToken,
		ApiAccessToken:      apiAccessToken,
		ApiRefreshToken:     apiRefreshToken,
	}, nil
}

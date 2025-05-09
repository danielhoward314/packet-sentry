package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/hashes"
	pbAdministrators "github.com/danielhoward314/packet-sentry/protogen/golang/administrators"
	"github.com/go-redis/redis/v8"
)

const (
	svcNameAdministrators = "administrators"
)

// administratorsService implements the administrators gRPC service
type administratorsService struct {
	pbAdministrators.UnimplementedAdministratorsServiceServer
	datastore             *dao.Datastore
	logger                *slog.Logger
	registrationDatastore dao.RegistrationDatastore
	smtpDialer            *gomail.Dialer
	webConsoleURL         string
}

func NewAdministratorsService(
	datastore *dao.Datastore,
	baseLogger *slog.Logger,
	registrationDatastore dao.RegistrationDatastore,
	smtpDialer *gomail.Dialer,
	webConsoleURL string,
) pbAdministrators.AdministratorsServiceServer {
	childLogger := baseLogger.With(slog.String("service", svcNameAdministrators))

	return &administratorsService{
		datastore:             datastore,
		logger:                childLogger,
		registrationDatastore: registrationDatastore,
		smtpDialer:            smtpDialer,
		webConsoleURL:         webConsoleURL,
	}
}

func (as *administratorsService) Create(ctx context.Context, request *pbAdministrators.CreateAdministratorRequest) (*pbAdministrators.Empty, error) {
	if request.OrganizationId == "" {
		as.logger.Error("invalid organization id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization id")
	}
	if request.Email == "" {
		as.logger.Error("invalid email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if request.DisplayName == "" {
		as.logger.Error("invalid display name")
		return nil, status.Errorf(codes.InvalidArgument, "invalid display name")
	}
	if request.AuthorizationRole == "" {
		as.logger.Error("invalid authorization role")
		return nil, status.Errorf(codes.InvalidArgument, "invalid authorization role")
	}
	administrator := &dao.Administrator{
		Email:             request.Email,
		DisplayName:       request.DisplayName,
		OrganizationID:    request.OrganizationId,
		AuthorizationRole: request.AuthorizationRole,
	}
	randomInitialPassword, err := generateInitialPassword()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	id, err := as.datastore.Administrators.Create(administrator, randomInitialPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	organization, err := as.datastore.Organizations.Read(request.OrganizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	as.logger.Info("creating registration data", "organization_id", request.OrganizationId, "administrator_id", id)
	token, emailCode, err := as.registrationDatastore.Create(&dao.Registration{
		OrganizationID:  request.OrganizationId,
		AdministratorID: id,
	})
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create registration")
	}
	activationLink := as.webConsoleURL + "/activate?token=" + token
	emailTemplateData := struct {
		ActivationLink   string
		Code             string
		OrganizationName string
	}{
		ActivationLink:   activationLink,
		Code:             emailCode,
		OrganizationName: organization.Name,
	}
	as.logger.Info("parsing new admin email template")
	tmpl, err := template.ParseFiles("templates/activate_new_admin.html")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse new admin template %s", err.Error())
	}
	var body bytes.Buffer
	as.logger.Info("executing new admin email template")
	err = tmpl.Execute(&body, emailTemplateData)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to executing new admin email template %s", err.Error())
	}

	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", request.Email)
	m.SetHeader("Subject", "Packet Sentry: Welcome")
	m.SetBody("text/html", body.String())
	as.logger.Info("sending new admin email")
	err = as.smtpDialer.DialAndSend(m)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to send new admin email %s", err.Error())
	}
	return &pbAdministrators.Empty{}, nil
}

func generateInitialPassword() (string, error) {
	const passwordLength = 24
	randomBytes := make([]byte, passwordLength)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	password := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
	return password, nil
}

func (as *administratorsService) Activate(ctx context.Context, request *pbAdministrators.ActivateAdministratorsRequest) (*pbAdministrators.Empty, error) {
	if request.Token == "" {
		as.logger.Error("invalid token")
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}
	if request.VerificationCode == "" {
		as.logger.Error("invalid verification code")
		return nil, status.Errorf(codes.InvalidArgument, "invalid verification code")
	}
	if request.Password == "" {
		as.logger.Error("invalid password")
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
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
	passwordHash, err := hashes.HashCleartextWithBCrypt(request.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save new admin's password")
	}
	administrator.PasswordHash = passwordHash
	err = as.datastore.Administrators.Update(administrator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update administrator data: %s", err.Error())
	}
	err = as.registrationDatastore.Delete(request.Token)
	if err != nil {
		// non-fatal error, the registration data has a short TTL
		as.logger.Warn("failed to delete registration data")
	}
	return &pbAdministrators.Empty{}, nil
}

func (as *administratorsService) Get(ctx context.Context, request *pbAdministrators.GetAdministratorRequest) (*pbAdministrators.GetAdministratorResponse, error) {
	if request.Id == "" {
		as.logger.Error("invalid id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid id")
	}
	administrator, err := as.datastore.Administrators.Read(request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}
	return &pbAdministrators.GetAdministratorResponse{
		Id:                administrator.ID,
		Email:             administrator.Email,
		DisplayName:       administrator.DisplayName,
		OrganizationId:    administrator.OrganizationID,
		Verified:          administrator.Verified,
		AuthorizationRole: administrator.AuthorizationRole,
	}, nil
}

func (as *administratorsService) List(ctx context.Context, request *pbAdministrators.ListAdministratorsRequest) (*pbAdministrators.ListAdministratorsResponse, error) {
	if request.OrganizationId == "" {
		as.logger.Error("invalid organization id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization id")
	}
	administrators, err := as.datastore.Administrators.List(request.OrganizationId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "no administrators found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to list administrators data: %s", err.Error())
	}
	response := pbAdministrators.ListAdministratorsResponse{
		Administrators: make([]*pbAdministrators.GetAdministratorResponse, 0, len(administrators)),
	}
	for _, daoAdministrator := range administrators {
		response.Administrators = append(response.Administrators, &pbAdministrators.GetAdministratorResponse{
			Id:                daoAdministrator.ID,
			Email:             daoAdministrator.Email,
			DisplayName:       daoAdministrator.DisplayName,
			OrganizationId:    daoAdministrator.OrganizationID,
			Verified:          daoAdministrator.Verified,
			AuthorizationRole: daoAdministrator.AuthorizationRole,
		})
	}
	return &response, nil
}

func (as *administratorsService) Update(ctx context.Context, request *pbAdministrators.UpdateAdministratorRequest) (*pbAdministrators.Empty, error) {
	if request.Id == "" {
		as.logger.Error("invalid id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid id")
	}
	administrator, err := as.datastore.Administrators.Read(request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "administrator not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to read administrator data: %s", err.Error())
	}

	hasAtLeastOneChange := false
	if request.AuthorizationRole != "" {
		administrator.AuthorizationRole = request.AuthorizationRole
		hasAtLeastOneChange = true
	}
	if request.DisplayName != "" {
		administrator.DisplayName = request.DisplayName
		hasAtLeastOneChange = true
	}
	if request.Email != "" {
		administrator.Email = request.Email
		hasAtLeastOneChange = true
	}
	if !hasAtLeastOneChange {
		as.logger.Error("administrator update must update at least one modifiable field")
		return nil, status.Errorf(codes.InvalidArgument, "administrator update must update at least one modifiable field")
	}
	err = as.datastore.Administrators.Update(administrator)
	if err != nil {
		as.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to update administrator")
	}
	return &pbAdministrators.Empty{}, nil
}

func (as *administratorsService) Delete(ctx context.Context, request *pbAdministrators.DeleteAdministratorRequest) (*pbAdministrators.Empty, error) {
	if request.Id == "" {
		as.logger.Error("invalid id")
		return nil, status.Errorf(codes.InvalidArgument, "invalid id")
	}
	rowsAffected, err := as.datastore.Administrators.Delete(request.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "no administrators found for delete: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete administrator data: %s", err.Error())
	}
	if rowsAffected != 1 {
		return nil, status.Errorf(
			codes.Internal,
			"invalid number of administrators rows deleted for delete by id %s: %d",
			request.Id,
			rowsAffected,
		)
	}
	return &pbAdministrators.Empty{}, nil
}

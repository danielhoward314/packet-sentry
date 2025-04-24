package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/danielhoward314/packet-sentry/dao"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
)

type bootstrapService struct {
	pbBootstrap.UnimplementedBootstrapServiceServer
	BaseLogger *slog.Logger
	CACert     *x509.Certificate
	CAKey      *rsa.PrivateKey
	Datastore  *dao.Datastore
	Logger     *slog.Logger
}

func NewBootstrapService(
	datastore *dao.Datastore,
	logger *slog.Logger,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) pbBootstrap.BootstrapServiceServer {
	return &bootstrapService{
		Datastore: datastore,
		CACert:    caCert,
		CAKey:     caKey,
		Logger:    logger,
	}
}

func (bs *bootstrapService) RequestCertificate(ctx context.Context, req *pbBootstrap.CertificateRequest) (*pbBootstrap.CertificateResponse, error) {
	logger := bs.Logger.With(psLog.KeyFunction, "bootstrapService.RequestCertificate")

	logger.Info("decoding CSR", psLog.KeyCertificateSigningRequest, req.Csr)
	block, _ := pem.Decode([]byte(req.Csr))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("bad CSR")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CSR")
	}
	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("bad CSR")
	}

	// DAO struct that will be used to create or update the device
	device := &dao.Device{
		OSUniqueIdentifier: csr.Subject.CommonName,
	}

	if req.IsRenewal {
		logger.Info("certificate request is renewal, checking cert fingerprint in request against persisted one", psLog.KeyExistingCertFingerprint, req.ExistingCertFingerprint)
		existingDevice, err := bs.Datastore.Devices.GetDeviceByOSUniqueIdentifier(csr.Subject.CommonName)
		if err != nil {
			logger.Error("error looking up device by os_unique_identifier", psLog.KeyError, err)
			return nil, fmt.Errorf("error looking up device by os_unique_identifier: %v", err)
		}
		if existingDevice == nil {
			logger.Error("device not found by os_unique_identifier")
			return nil, fmt.Errorf("cannot renew cert for device not found by os_unique_identifier: %s", csr.Subject.CommonName)
		}
		if strings.TrimSpace(strings.ToLower(existingDevice.ClientCertFingerprint)) != strings.TrimSpace(strings.ToLower(req.ExistingCertFingerprint)) {
			logger.Error("client cert fingerprint in request does not match persisted one")
			return nil, fmt.Errorf("client cert fingerprint in request does not match persisted one")
		}
		device.ID = existingDevice.ID
		device.OrganizationID = existingDevice.OrganizationID
		device.InterfaceBPFAssociations = existingDevice.InterfaceBPFAssociations
		device.PreviousAssociations = existingDevice.PreviousAssociations
	} else {
		logger.Info("certificate request is not renewal, checking install key in request against persisted one")
		if req.InstallKey == "" {
			return nil, fmt.Errorf("missing install key for first certificate request")
		}
		validatedKey, err := bs.Datastore.InstallKeys.Validate(req.InstallKey)
		if err != nil {
			return nil, fmt.Errorf("error validating install key: %v", err)
		}
		if validatedKey == nil || validatedKey.OrganizationID == "" || validatedKey.AdministratorID == "" {
			return nil, fmt.Errorf("invalid install key")
		}
		device.OrganizationID = validatedKey.OrganizationID
		rowsDeleted, err := bs.Datastore.InstallKeys.Delete(req.InstallKey)
		if err != nil {
			logger.Warn("error deleting install key", psLog.KeyError, err)
		} else if rowsDeleted == 0 {
			logger.Warn("no install key deleted after validation")
		}
	}

	logger.Info("generating certificate serial number")
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		logger.Error("error generating serial number", psLog.KeyError, err)
		return nil, fmt.Errorf("error generating serial number %v", err)
	}

	// Generate new cert from CSR
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      csr.Subject,
		NotBefore:    time.Now().Add(-1 * time.Minute),
		// NotAfter:     time.Now().AddDate(1, 0, 0), // 1 year
		NotAfter:              time.Now().Add(10 * time.Minute),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	logger.Info("creating certificate from CSR")
	certDER, err := x509.CreateCertificate(rand.Reader, template, bs.CACert, csr.PublicKey, bs.CAKey)
	if err != nil {
		logger.Error("error creating certificate from CSR", psLog.KeyError, err)
		return nil, fmt.Errorf("bad CSR")
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	fingerprint := sha256.Sum256(certDER)
	newCertFingerprint := hex.EncodeToString(fingerprint[:])
	logger.Info("new cert fingerprint", psLog.KeyCertFingerprint, newCertFingerprint)
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bs.CACert.Raw})

	if req.IsRenewal {
		logger.Info("updating device with new client cert pem and cert fingerprint")
		device.ClientCertPEM = string(certPEM)
		device.ClientCertFingerprint = strings.TrimSpace(strings.ToLower(newCertFingerprint))
		err = bs.Datastore.Devices.Update(device)
		if err != nil {
			return nil, fmt.Errorf("error updating device: %v", err)
		}
	} else {
		logger.Info("creating device with new client cert pem and cert fingerprint")
		device.ClientCertPEM = string(certPEM)
		device.ClientCertFingerprint = strings.TrimSpace(strings.ToLower(newCertFingerprint))
		err = bs.Datastore.Devices.Create(device)
		if err != nil {
			return nil, fmt.Errorf("error creating device: %v", err)
		}
	}

	return &pbBootstrap.CertificateResponse{
		ClientCertificate:     string(certPEM),
		CaCertificate:         string(caCertPEM),
		ClientCertFingerprint: newCertFingerprint,
	}, nil

}

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
	"sync"
	"time"

	"github.com/danielhoward314/packet-sentry/dao"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
)

// TODO: get rid of these in-memory values used for PoC; replace with persisted certificates
var (
	issuedCert struct {
		Fingerprint string
		CertPEM     []byte
		sync.Mutex
	}
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

	// TODO: obviously replace the in-memory stuff with persistence
	issuedCert.Lock()
	defer issuedCert.Unlock()

	if req.IsRenewal {
		logger.Info("certificate request is renewal, checking cert fingerprint in request against persisted one")
		// TODO: check in database instead, should be able to use the `csr.Subject.CommonName` for lookup
		finalReqFingerprint := strings.ToLower(strings.TrimSpace(req.ExistingCertFingerprint))
		finalIssuedCertFingerprint := strings.ToLower(strings.TrimSpace(issuedCert.Fingerprint))
		fmt.Printf("finalReqFingerprint: \n\n%s\n\n", finalReqFingerprint)
		fmt.Printf("finalIssuedCertFingerprint: \n\n%s\n\n", finalIssuedCertFingerprint)
		if finalReqFingerprint != finalIssuedCertFingerprint {
			return nil, fmt.Errorf("mismatch of existing cert fingerprint in request versus the one in server memory")
		}
	} else {
		logger.Info("certificate request is not renewal, checking install key in request against persisted one")
		if req.InstallKey == "" {
			return nil, fmt.Errorf("missing install key for first certificate request")
		}
		err = bs.Datastore.InstallKeys.Validate(req.InstallKey)
		if err != nil {
			return nil, fmt.Errorf("error validating install key: %v", err)
		}
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
		return nil, fmt.Errorf("bad CSR")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	fingerprint := sha256.Sum256(certDER)
	issuedCert.Fingerprint = hex.EncodeToString(fingerprint[:])
	issuedCert.CertPEM = certPEM

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bs.CACert.Raw})

	return &pbBootstrap.CertificateResponse{
		ClientCertificate:     string(certPEM),
		CaCertificate:         string(caCertPEM),
		ClientCertFingerprint: issuedCert.Fingerprint,
	}, nil

}

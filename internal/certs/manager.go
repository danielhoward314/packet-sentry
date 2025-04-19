package certs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/config"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psOS "github.com/danielhoward314/packet-sentry/internal/os"
	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
)

const (
	logAttrValSvcName              = "certificateManager"
	packetSentryIssuerCN           = "Packet Sentry"
	packetSentrySubjectOrg         = "Packet Sentry"
	pemBlockTypeCertificate        = "CERTIFICATE"
	pemBlockTypeCertificateRequest = "CERTIFICATE REQUEST"
	pemBlockTypeRSAPrivateKey      = "RSA PRIVATE KEY"
)

// CertificateManager is the manager interface for certificate and mTLS client operations
type CertificateManager interface {
	Init() error
	Start()
	Stop()
}

type certificateManager struct {
	agentMTLSClientBroadcaster *broadcast.AgentMTLSClientBroadcaster
	agentMTLSClientTarget      string
	bootstrapClient            pbBootstrap.BootstrapServiceClient
	caCert                     *x509.Certificate
	cancelFunc                 context.CancelFunc
	certCheckInterval          time.Duration
	clientCert                 *x509.Certificate
	clientPrivKey              *rsa.PrivateKey
	ctx                        context.Context
	logger                     *slog.Logger
	stopOnce                   sync.Once
	systemInfo                 psOS.SystemInfo
}

// NewCertificateManager returns an implementation of the CertificateManager interface
func NewCertificateManager(
	ctx context.Context,
	baseLogger *slog.Logger,
	systemInfo psOS.SystemInfo,
	boostrapClient pbBootstrap.BootstrapServiceClient,
	agentMTLSClientBroadcaster *broadcast.AgentMTLSClientBroadcaster,
	agentMTLSClientTarget string,
) CertificateManager {
	childCtx, cancelFunc := context.WithCancel(ctx)
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))

	return &certificateManager{
		agentMTLSClientBroadcaster: agentMTLSClientBroadcaster,
		agentMTLSClientTarget:      agentMTLSClientTarget,
		bootstrapClient:            boostrapClient,
		cancelFunc:                 cancelFunc,
		certCheckInterval:          config.GetCertCheckInterval(),
		ctx:                        childCtx,
		logger:                     childLogger,
		systemInfo:                 systemInfo,
	}
}

// CertExpiringSoonError is a custom error type signaling that a cert is close to expiration.
type CertExpiringSoonError struct {
	NotAfter        time.Time
	DaysUntilExpiry float64
}

// Error implements the Error interface for the CertExpiringSoonError
func (e *CertExpiringSoonError) Error() string {
	return fmt.Sprintf("certificate is expiring soon: in %.0f days on %s", e.DaysUntilExpiry, e.NotAfter)
}

// Bootstrap represents the contents of the agentBootstrap.json file
type Bootstrap struct {
	InstallKey string `json:"installKey"`
}

// Init is called during startup to create and publish an mTLS client.
// It does this by either:
//
//	requesting the first-time client cert
//	renewing the cert when near expiry
//	using the existing cert on disk when not near expiry
func (cm *certificateManager) Init() error {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.Init")
	err := cm.hasValidCert()
	if err != nil {
		isRenewal := false
		switch err.(type) {
		case *CertExpiringSoonError:
			logger.Warn("client certificate will expire within 30 days, requesting renewal")
			isRenewal = true
		default:
			logger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
		}
		err = cm.requestCert(isRenewal)
		if err != nil {
			logger.Error("failed to get client certificate from server", psLog.KeyError, err)
			return err
		}
	}

	mTLSClientConn, err := cm.createMTLSConnection(cm.agentMTLSClientTarget)
	if err != nil {
		logger.Error("failed to get mTLS client connection", psLog.KeyError, err)
		return err
	}
	cm.agentMTLSClientBroadcaster.Publish(mTLSClientConn)

	return nil
}

// Start is the certManager goroutine that periodically checks cert validity
// and performs similar work to Init, except it can skip creating and publishing an mTLS client
// when the existing cert is still valid
func (cm *certificateManager) Start() {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.Start")
	logger.Info("starting certificate manager")
	for {
		select {
		case <-time.After(cm.certCheckInterval):
			logger.Info("cert check interval elapsed, checking validity of existing cert")
			err := cm.hasValidCert()
			// unlike Init, which must publish an mTLS client as part of the startup sequence,
			// Start can skip the rest of this work when we have a valid cert
			if err == nil {
				continue
			}

			isRenewal := false
			switch err.(type) {
			case *CertExpiringSoonError:
				logger.Warn("client certificate will expire within 30 days, requesting renewal")
				isRenewal = true
			default:
				logger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
			}
			err = cm.requestCert(isRenewal)
			if err != nil {
				logger.Error("failed to get client certificate from server", psLog.KeyError, err)
				continue
			}

			mTLSClientConn, err := cm.createMTLSConnection(cm.agentMTLSClientTarget)
			if err != nil {
				logger.Error("failed to get mTLS client connection", psLog.KeyError, err)
				continue
			}
			cm.agentMTLSClientBroadcaster.Publish(mTLSClientConn)
		case <-cm.ctx.Done():
			logger.Error("certificate manager context canceled")
			return
		}
	}
}

func (cm *certificateManager) Stop() {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.Stop")

	cm.stopOnce.Do(func() {
		logger.Info("stopping certificate manager")
		cm.cancelFunc()
	})
}

func (cm *certificateManager) getCertFromDisk(filePath string) (*x509.Certificate, error) {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.getCertFromDisk")

	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	logger.Info("decoding certificate PEM blocks")
	pemBlocks := make([]*pem.Block, 0)
	for {
		pemBlock, rest := pem.Decode(pemBytes)
		if pemBlock == nil {
			return nil, fmt.Errorf("failed to decode PEM block of certificate")
		}
		pemBlocks = append(pemBlocks, pemBlock)
		if len(rest) == 0 {
			break
		}
	}
	logger.Info("validating certificate PEM block")
	if len(pemBlocks) > 1 {
		return nil, fmt.Errorf("found more than one PEM block in certificate")
	}
	if pemBlocks[0].Type != pemBlockTypeCertificate {
		return nil, fmt.Errorf("found incorrect PEM block type in certificate")
	}
	logger.Info("parsing certificate PEM block")
	cert, err := x509.ParseCertificate(pemBlocks[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM bytes of certificate")
	}

	return cert, nil
}

func (cm *certificateManager) getPrivateKeyFromDisk(filepath string) (*rsa.PrivateKey, error) {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.getPrivateKeyFromDisk")

	logger.Info("reading client private key from disk")
	pemBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	logger.Info("decoding client private key PEM blocks")
	pemBlocks := make([]*pem.Block, 0)
	for {
		pemBlock, rest := pem.Decode(pemBytes)
		if pemBlock == nil {
			return nil, fmt.Errorf("failed to decode PEM block of client private key")
		}
		pemBlocks = append(pemBlocks, pemBlock)
		if len(rest) == 0 {
			break
		}
	}
	logger.Info("validating client private key PEM block")
	if len(pemBlocks) > 1 {
		return nil, fmt.Errorf("found more than one PEM block in client private key")
	}
	if pemBlocks[0].Type != pemBlockTypeRSAPrivateKey {
		return nil, fmt.Errorf("found incorrect PEM block type in client private key")
	}
	logger.Info("parsing client private key PEM block")
	privKey, err := x509.ParsePKCS1PrivateKey(pemBlocks[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM bytes of client private key")
	}
	return privKey, nil
}

func (cm *certificateManager) hasValidCert() error {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.hasValidCert")

	if cm.clientCert == nil {
		logger.Info("client cert not in-memory, reading client certificate from disk")
		cert, err := cm.getCertFromDisk(config.GetClientCertFilePath())
		if err != nil {
			return fmt.Errorf("failed to get client certificate from disk")
		}
		cm.clientCert = cert
	} else {
		logger.Info("client cert is already in-memory, skipping reading client certificate from disk")
	}

	logger.Info("validating client certificate NotBefore and NotAfter")
	now := time.Now()
	if now.Before(cm.clientCert.NotBefore) {
		return fmt.Errorf("current timestamp is before cert's NotBefore: %s", cm.clientCert.NotBefore)
	}
	if now.After(cm.clientCert.NotAfter) {
		return fmt.Errorf("current timestamp is after cert's NotAfter: %s", cm.clientCert.NotAfter)
	}

	logger.Info("checking if client certificate is within 30 days of expiration")
	daysUntilExpiry := cm.clientCert.NotAfter.Sub(now).Hours() / 24
	if daysUntilExpiry <= 30 && daysUntilExpiry > 0 {
		return &CertExpiringSoonError{
			NotAfter:        cm.clientCert.NotAfter,
			DaysUntilExpiry: daysUntilExpiry,
		}
	}
	return nil
}

func (cm *certificateManager) requestCert(isRenewal bool) error {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.requestCert")

	var csrTemplate *x509.CertificateRequest
	var privKey *rsa.PrivateKey
	var err error

	if isRenewal {
		logger.Info("renewing client certificate")
		if cm.clientCert == nil {
			logger.Info("client cert not in-memory, reading from disk")
			cert, err := cm.getCertFromDisk(config.GetClientCertFilePath())
			if err != nil {
				return fmt.Errorf("failed to get client certificate from disk")
			}
			cm.clientCert = cert
		}
		privKey, err = cm.getPrivateKeyFromDisk(config.GetPrivateKeyFilePath())
		if err != nil {
			return fmt.Errorf("failed to get client private key from disk")
		}
		cm.clientPrivKey = privKey
	} else {
		logger.Info("requesting first client certificate, creating private key")
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		privKeyPEM := &pem.Block{
			Type:  pemBlockTypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		}
		privKeyPEMBytes := pem.EncodeToMemory(privKeyPEM)
		logger.Info("writing private key to disk")
		err = os.WriteFile(config.GetPrivateKeyFilePath(), privKeyPEMBytes, 0o600)
		if err != nil {
			return err
		}
		cm.clientPrivKey = privKey
	}

	logger.Info("getting unique system identifier")
	uniqueSystemID, err := cm.systemInfo.GetUniqueSystemIdentifier()
	if err != nil {
		return err
	}

	csrTemplate = &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   uniqueSystemID,
			Organization: []string{packetSentrySubjectOrg},
		},
	}

	logger.Info("creating CSR from template and private key")
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privKey)
	if err != nil {
		return err
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: pemBlockTypeCertificateRequest, Bytes: csrDER})

	req := &pbBootstrap.CertificateRequest{
		Csr:       string(csrPEM),
		IsRenewal: isRenewal,
	}

	if isRenewal {
		logger.Info("calculating fingerprint of existing cert for inclusion in cert renewal request")
		fp := sha256.Sum256(cm.clientCert.Raw)
		req.ExistingCertFingerprint = fmt.Sprintf("%X", fp[:])
	} else {
		logger.Info("reading bootstrap file for setting install key header for first cert request")
		bootstrapPath := config.GetBootstrapFilePath()
		content, err := os.ReadFile(bootstrapPath)
		if err != nil {
			return err
		}
		var bs Bootstrap
		if err := json.Unmarshal(content, &bs); err != nil {
			return fmt.Errorf("failed to parse bootstrap file: %w", err)
		}

		installKey := strings.TrimSpace(bs.InstallKey)
		if installKey == "" {
			return fmt.Errorf("install key needed to request first client certificate")
		}
		logger.Info("setting x-install-key in request")
		req.InstallKey = installKey
	}

	res, err := cm.bootstrapClient.RequestCertificate(cm.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get response for certificate requests %w", err)
	}

	logger.Info("decoding PEM block of client cert in response")
	newCertPEMBlock, _ := pem.Decode([]byte(res.ClientCertificate))
	if newCertPEMBlock == nil {
		return fmt.Errorf("failed to decode PEM block of certificate in response")
	}

	if newCertPEMBlock.Type != pemBlockTypeCertificate {
		return fmt.Errorf("expected CERTIFICATE PEM block, got %s", newCertPEMBlock.Type)
	}

	logger.Info("parsing client cert in response")
	newCert, err := x509.ParseCertificate(newCertPEMBlock.Bytes)
	if err != nil {
		return err
	}
	cm.clientCert = newCert

	logger.Info("decoding PEM block of CA cert in response")
	newCACertPEMBlock, _ := pem.Decode([]byte(res.CaCertificate))
	if newCACertPEMBlock == nil {
		return fmt.Errorf("failed to decode PEM block of certificate in response")
	}

	if newCACertPEMBlock.Type != pemBlockTypeCertificate {
		return fmt.Errorf("expected CERTIFICATE PEM block, got %s", newCACertPEMBlock.Type)
	}

	logger.Info("parsing CA cert in response")
	newCACert, err := x509.ParseCertificate(newCACertPEMBlock.Bytes)
	if err != nil {
		return err
	}
	cm.caCert = newCACert

	clientCertFilePath := config.GetClientCertFilePath()
	caCertFilePath := config.GetCACertFilePath()

	if isRenewal {
		logger.Info("deleting old client certificate")
		err = os.Remove(clientCertFilePath)
		if err != nil {
			return err
		}
		logger.Info("deleting old CA certificate")
		err = os.Remove(caCertFilePath)
		if err != nil {
			return err
		}
	}

	logger.Info("writing client certificate to disk")
	// the `clientCertificate` field of the cert response is already PEM bytes
	err = os.WriteFile(clientCertFilePath, []byte(res.ClientCertificate), 0o600)
	if err != nil {
		return err
	}

	logger.Info("writing CA certificate to disk")
	// the `caCertificate` field of the cert response is already PEM bytes
	err = os.WriteFile(caCertFilePath, []byte(res.CaCertificate), 0o600)
	if err != nil {
		return err
	}

	return nil
}

func (cm *certificateManager) createMTLSConnection(addr string) (*grpc.ClientConn, error) {
	logger := cm.logger.With(psLog.KeyFunction, "CertificateManager.createMTLSConnection")

	logger.Info("creating mTLS gRPC connection")

	if cm.clientCert == nil {
		clientCert, err := cm.getCertFromDisk(config.GetClientCertFilePath())
		if err != nil {
			return nil, fmt.Errorf("failed to create mTLS client due to missing client cert, private key, or CA cert")
		}
		cm.clientCert = clientCert
	}
	if cm.clientPrivKey == nil {
		clientPrivKey, err := cm.getPrivateKeyFromDisk(config.GetPrivateKeyFilePath())
		if err != nil {
			return nil, fmt.Errorf("failed to create mTLS client due to missing client cert, private key, or CA cert")
		}
		cm.clientPrivKey = clientPrivKey
	}
	if cm.caCert == nil {
		caCert, err := cm.getCertFromDisk(config.GetCACertFilePath())
		if err != nil {
			return nil, fmt.Errorf("failed to create mTLS client due to missing client cert, private key, or CA cert")
		}
		cm.caCert = caCert
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cm.clientCert.Raw},
		PrivateKey:  cm.clientPrivKey,
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(cm.caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		RootCAs:            rootCAs,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		logger.Error("failed to create grpc connection", psLog.KeyError, err)
		return nil, err
	}

	logger.Info("successfully created gRPC mTLS connection")
	return conn, nil
}

package certs

import (
	"bytes"
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
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/danielhoward314/packet-sentry/internal/broadcast"
	"github.com/danielhoward314/packet-sentry/internal/config"
	psLog "github.com/danielhoward314/packet-sentry/internal/log"
	psOS "github.com/danielhoward314/packet-sentry/internal/os"
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
	caCert                *x509.Certificate
	certCheckInterval     time.Duration
	clientCert            *x509.Certificate
	clientPrivKey         *rsa.PrivateKey
	ctx                   context.Context
	logger                *slog.Logger
	mTLSClientBroadcaster *broadcast.MTLSClientBroadcaster
	shutdownChannel       chan struct{}
	systemInfo            psOS.SystemInfo
	wg                    sync.WaitGroup
}

// NewCertificateManager returns an implementation of the CertificateManager interface
func NewCertificateManager(ctx context.Context, baseLogger *slog.Logger, systemInfo psOS.SystemInfo, mtlsClientBroadCaster *broadcast.MTLSClientBroadcaster) CertificateManager {
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))
	return &certificateManager{
		certCheckInterval:     config.GetCertCheckInterval(),
		ctx:                   ctx,
		logger:                childLogger,
		mTLSClientBroadcaster: mtlsClientBroadCaster,
		shutdownChannel:       make(chan struct{}),
		systemInfo:            systemInfo,
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

// CertificateRequest represents the request body sent to the `/certificates` endpoint
// issued for first-time and renewal client certificate requests
type CertificateRequest struct {
	CSR                     string `json:"csr"`
	ExistingCertFingerprint string `json:"existingCertFingerprint,omitempty"`
}

// CertificateResponse represents the expected response body for `/certificates` requests
type CertificateResponse struct {
	ClientCertificate string `json:"clientCertificate"` // PEM-encoded certificate bytes
	CACertificate     string `json:"caCertificate"`     // PEM-encoded certificate bytes
}

func (cm *certificateManager) getCertFromDisk(filePath string) (*x509.Certificate, error) {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.getCertFromDisk")

	pemBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cm.logger.Info("decoding certificate PEM blocks")
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
	cm.logger.Info("validating certificate PEM block")
	if len(pemBlocks) > 1 {
		return nil, fmt.Errorf("found more than one PEM block in certificate")
	}
	if pemBlocks[0].Type != pemBlockTypeCertificate {
		return nil, fmt.Errorf("found incorrect PEM block type in certificate")
	}
	cm.logger.Info("parsing certificate PEM block")
	cert, err := x509.ParseCertificate(pemBlocks[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM bytes of certificate")
	}

	return cert, nil
}

func (cm *certificateManager) getPrivateKeyFromDisk(filepath string) (*rsa.PrivateKey, error) {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.getPrivateKeyFromDisk")

	cm.logger.Info("reading client private key from disk")
	pemBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	cm.logger.Info("decoding client private key PEM blocks")
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
	cm.logger.Info("validating client private key PEM block")
	if len(pemBlocks) > 1 {
		return nil, fmt.Errorf("found more than one PEM block in client private key")
	}
	if pemBlocks[0].Type != pemBlockTypeRSAPrivateKey {
		return nil, fmt.Errorf("found incorrect PEM block type in client private key")
	}
	cm.logger.Info("parsing client private key PEM block")
	privKey, err := x509.ParsePKCS1PrivateKey(pemBlocks[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM bytes of client private key")
	}
	return privKey, nil
}

func (cm *certificateManager) hasValidCert() error {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.hasValidCert")

	if cm.clientCert == nil {
		cm.logger.Info("client cert not in-memory, reading client certificate from disk")
		cert, err := cm.getCertFromDisk(config.GetClientCertFilePath())
		if err != nil {
			return fmt.Errorf("failed to get client certificate from disk")
		}
		cm.clientCert = cert
	} else {
		cm.logger.Info("client cert is already in-memory, skipping reading client certificate from disk")
	}

	cm.logger.Info("validating client certificate NotBefore and NotAfter")
	now := time.Now()
	if now.Before(cm.clientCert.NotBefore) {
		return fmt.Errorf("current timestamp is before cert's NotBefore: %s", cm.clientCert.NotBefore)
	}
	if now.After(cm.clientCert.NotAfter) {
		return fmt.Errorf("current timestamp is after cert's NotAfter: %s", cm.clientCert.NotAfter)
	}

	cm.logger.Info("checking if client certificate is withing 30 days of expiration")
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
	cm.logger.With(psLog.KeyFunction, "CertificateManager.requestCert")

	var csrTemplate *x509.CertificateRequest
	var privKey *rsa.PrivateKey
	var err error

	if isRenewal {
		cm.logger.Info("renewing client certificate")
		if cm.clientCert == nil {
			cm.logger.Info("client cert not in-memory, reading from disk")
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
		cm.logger.Info("requesting first client certificate, creating private key")
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		privKeyPEM := &pem.Block{
			Type:  pemBlockTypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		}
		privKeyPEMBytes := pem.EncodeToMemory(privKeyPEM)
		cm.logger.Info("writing private key to disk")
		err = os.WriteFile(config.GetPrivateKeyFilePath(), privKeyPEMBytes, 0o600)
		if err != nil {
			return err
		}
		cm.clientPrivKey = privKey
	}

	cm.logger.Info("getting unique system identifier")
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

	cm.logger.Info("creating CSR from template and private key")
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privKey)
	if err != nil {
		return err
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: pemBlockTypeCertificateRequest, Bytes: csrDER})

	reqBody := &CertificateRequest{
		CSR: string(csrPEM),
	}

	method := http.MethodPost
	// TODO: env-based handling of baseURL and, for local dev, either `InsecureSkipVerify: true` or putting server cert in tls config
	// also, the port should be the tls (not mTLS) server port
	url := "https://localhost:8443/certificates"
	client := http.DefaultClient
	// delete these lines when TODO above adds env-based config
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	client.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	if isRenewal {
		cm.logger.Info("calculating fingerprint of existing cert for inclusion in cert renewal request")
		fp := sha256.Sum256(cm.clientCert.Raw)
		reqBody.ExistingCertFingerprint = fmt.Sprintf("%X", fp[:])
		method = http.MethodPut
	}

	cm.logger.Info("marshaling request body")
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	cm.logger.Info("creating http request")
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if !isRenewal {
		cm.logger.Info("reading bootstrap file for setting install key header for first cert request")
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
		cm.logger.Info("setting x-install-key header")
		req.Header.Set("x-install-key", installKey)
	}

	cm.logger.Info("making http request")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200: %s", resp.Status)
	}

	cm.logger.Info("reading http response body")
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cm.logger.Info("unmarshaling http response body")
	var certResponse CertificateResponse
	err = json.Unmarshal(resBody, &certResponse)
	if err != nil {
		return err
	}

	cm.logger.Info("decoding PEM block of client cert in response")
	newCertPEMBlock, _ := pem.Decode([]byte(certResponse.ClientCertificate))
	if newCertPEMBlock == nil {
		return fmt.Errorf("failed to decode PEM block of certificate in http response")
	}

	if newCertPEMBlock.Type != pemBlockTypeCertificate {
		return fmt.Errorf("expected CERTIFICATE PEM block, got %s", newCertPEMBlock.Type)
	}

	cm.logger.Info("parsing client cert in response")
	newCert, err := x509.ParseCertificate(newCertPEMBlock.Bytes)
	if err != nil {
		return err
	}
	cm.clientCert = newCert

	cm.logger.Info("decoding PEM block of CA cert in response")
	newCACertPEMBlock, _ := pem.Decode([]byte(certResponse.CACertificate))
	if newCACertPEMBlock == nil {
		return fmt.Errorf("failed to decode PEM block of certificate in http response")
	}

	if newCACertPEMBlock.Type != pemBlockTypeCertificate {
		return fmt.Errorf("expected CERTIFICATE PEM block, got %s", newCACertPEMBlock.Type)
	}

	cm.logger.Info("parsing CA cert in response")
	newCACert, err := x509.ParseCertificate(newCACertPEMBlock.Bytes)
	if err != nil {
		return err
	}
	cm.caCert = newCACert

	clientCertFilePath := config.GetClientCertFilePath()
	caCertFilePath := config.GetCACertFilePath()

	if isRenewal {
		cm.logger.Info("deleting old client certificate")
		err = os.Remove(clientCertFilePath)
		if err != nil {
			return err
		}
		cm.logger.Info("deleting old CA certificate")
		err = os.Remove(caCertFilePath)
		if err != nil {
			return err
		}
	}

	cm.logger.Info("writing client certificate to disk")
	// the `clientCertificate` field of the cert response is already PEM bytes
	err = os.WriteFile(clientCertFilePath, []byte(certResponse.ClientCertificate), 0o600)
	if err != nil {
		return err
	}

	cm.logger.Info("writing CA certificate to disk")
	// the `caCertificate` field of the cert response is already PEM bytes
	err = os.WriteFile(caCertFilePath, []byte(certResponse.CACertificate), 0o600)
	if err != nil {
		return err
	}

	return nil
}

func (cm *certificateManager) getMTLSClient() (*http.Client, error) {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.getMTLSClient")

	cm.logger.Info("instantiating mTLS client")
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
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		RootCAs:            rootCAs,
	}

	mTLSClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	cm.logger.Info("testing mTLS communication with server")
	// TODO: env-based config of base URL
	res, err := mTLSClient.Get("https://localhost:9443/mtls-test")
	if err != nil {
		cm.logger.Error("mTLS request failed", psLog.KeyError, err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("mTLS response status is non-200: %d", res.StatusCode)
		cm.logger.Error("failed mTLS readiness check", psLog.KeyError, err)
		return nil, err
	}

	cm.logger.Info("mTLS client and server communication successful")
	return mTLSClient, nil
}

// Init is called during startup to create and publish an mTLS client.
// It does this by either:
//
//	requesting the first-time client cert
//	renewing the cert when near expiry
//	using the existing cert on disk when not near expiry
func (cm *certificateManager) Init() error {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.Init")
	err := cm.hasValidCert()
	if err != nil {
		isRenewal := false
		switch err.(type) {
		case *CertExpiringSoonError:
			cm.logger.Warn("client certificate will expire within 30 days, requesting renewal")
			isRenewal = true
		default:
			cm.logger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
		}
		err = cm.requestCert(isRenewal)
		if err != nil {
			cm.logger.Error("failed to get client certificate from server", psLog.KeyError, err)
			return err
		}
	}

	mTLSClient, err := cm.getMTLSClient()
	if err != nil {
		cm.logger.Error("failed to get mTLS client", psLog.KeyError, err)
		return err
	}
	cm.mTLSClientBroadcaster.Publish(mTLSClient)

	return nil
}

// Start is the certManager goroutine that periodically checks cert validity
// and performs similar work to Init, except it can skip creating and publishing an mTLS client
// when the existing cert is still valid
func (cm *certificateManager) Start() {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.Start")
	cm.logger.Info("starting certificate manager")

	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		for {
			select {
			case <-time.After(cm.certCheckInterval):
				cm.logger.Info("cert check interval elapsed, checking validity of existing cert")
				err := cm.hasValidCert()
				// unlike Init, which must publish an mTLS client as part of the startup sequence,
				// Start can skip the rest of this work when we have a valid cert
				if err == nil {
					continue
				}

				isRenewal := false
				switch err.(type) {
				case *CertExpiringSoonError:
					cm.logger.Warn("client certificate will expire within 30 days, requesting renewal")
					isRenewal = true
				default:
					cm.logger.Warn("failed to find client certificate on disk, assuming this is first client cert request")
				}
				err = cm.requestCert(isRenewal)
				if err != nil {
					cm.logger.Error("failed to get client certificate from server", psLog.KeyError, err)
					continue
				}

				mTLSClient, err := cm.getMTLSClient()
				if err != nil {
					cm.logger.Error("failed to get mTLS client", psLog.KeyError, err)
					continue
				}
				cm.mTLSClientBroadcaster.Publish(mTLSClient)
			case <-cm.shutdownChannel:
				return
			}
		}
	}()
}

// Stop closes the shutdown channel and blocks until wait group is Done.
// The Start goroutine reacts to closed shutdown channel by returning,
// which triggers the deferred `cm.wg.Done()` call and unblocks Stop.
func (cm *certificateManager) Stop() {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.Stop")
	cm.logger.Info("stopping certificate manager")
	close(cm.shutdownChannel)
	cm.wg.Wait()
}

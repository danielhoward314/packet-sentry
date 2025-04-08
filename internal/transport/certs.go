package transport

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
	"time"

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

type CertificateManager interface {
	HasValidCert() error
	RequestCert(isRenewal bool) error
	GetMTLSClient() (*http.Client, error)
}

type certificateManager struct {
	ctx           context.Context
	caCert        *x509.Certificate
	clientCert    *x509.Certificate
	clientPrivKey *rsa.PrivateKey
	logger        *slog.Logger
	systemInfo    psOS.SystemInfo
}

func NewCertificateManager(ctx context.Context, baseLogger *slog.Logger, systemInfo psOS.SystemInfo) CertificateManager {
	childLogger := baseLogger.With(slog.String(psLog.KeyServiceName, logAttrValSvcName))
	return &certificateManager{
		ctx:        ctx,
		logger:     childLogger,
		systemInfo: systemInfo,
	}
}

// CertExpiringSoonError represents a warning that a cert is close to expiration.
type CertExpiringSoonError struct {
	NotAfter        time.Time
	DaysUntilExpiry float64
}

func (e *CertExpiringSoonError) Error() string {
	return fmt.Sprintf("certificate is expiring soon: in %.0f days on %s", e.DaysUntilExpiry, e.NotAfter)
}

type Bootstrap struct {
	InstallKey string `json:"installKey"`
}

type CertificateRequest struct {
	CSR                     string `json:"csr"`
	ExistingCertFingerprint string `json:"existingCertFingerprint,omitempty"`
}

type CertificateResponse struct {
	ClientCertificate string `json:"clientCertificate"` // PEM-encoded certificate bytes
	CACertificate     string `json:"caCertificate"`     // PEM-encoded certificate bytes
}

func (cm *certificateManager) getCertFromDisk() (*x509.Certificate, error) {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.getCertFromDisk")

	pemBytes, err := os.ReadFile(config.GetClientCertFilePath())
	if err != nil {
		return nil, err
	}

	cm.logger.Info("decoding client certificate PEM blocks")
	pemBlocks := make([]*pem.Block, 0)
	for {
		pemBlock, rest := pem.Decode(pemBytes)
		if pemBlock == nil {
			return nil, fmt.Errorf("failed to decode PEM block of client certificate")
		}
		pemBlocks = append(pemBlocks, pemBlock)
		if len(rest) == 0 {
			break
		}
	}
	cm.logger.Info("validating client certificate PEM block")
	if len(pemBlocks) > 1 {
		return nil, fmt.Errorf("found more than one PEM block in client certificate")
	}
	if pemBlocks[0].Type != pemBlockTypeCertificate {
		return nil, fmt.Errorf("found incorrect PEM block type in client certificate")
	}
	cm.logger.Info("parsing client certificate PEM block")
	cert, err := x509.ParseCertificate(pemBlocks[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM bytes of client certificate")
	}

	return cert, nil
}

func (cm *certificateManager) HasValidCert() error {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.HasValidCert")

	if cm.clientCert == nil {
		cm.logger.Info("client cert not in-memory, reading client certificate from disk")
		cert, err := cm.getCertFromDisk()
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

func (cm *certificateManager) RequestCert(isRenewal bool) error {
	cm.logger.With(psLog.KeyFunction, "CertificateManager.RequestCert")

	cm.logger.Info("getting unique system identifier")
	uniqueSystemID, err := cm.systemInfo.GetUniqueSystemIdentifier()
	if err != nil {
		return err
	}

	var csrTemplate *x509.CertificateRequest
	var privKey *rsa.PrivateKey

	if isRenewal {
		cm.logger.Info("renewing client certificate")
		// should already be in-memory from previous HasValidCert call
		if cm.clientCert == nil {
			cm.logger.Info("client cert not in-memory, reading from disk")
			cert, err := cm.getCertFromDisk()
			if err != nil {
				return fmt.Errorf("failed to get client certificate from disk")
			}
			cm.clientCert = cert
		}

		cm.logger.Info("reading client private key from disk")
		pemBytes, err := os.ReadFile(config.GetPrivateKeyFilePath())
		if err != nil {
			return err
		}

		cm.logger.Info("decoding client private key PEM blocks")
		pemBlocks := make([]*pem.Block, 0)
		for {
			pemBlock, rest := pem.Decode(pemBytes)
			if pemBlock == nil {
				return fmt.Errorf("failed to decode PEM block of client private key")
			}
			pemBlocks = append(pemBlocks, pemBlock)
			if len(rest) == 0 {
				break
			}
		}
		cm.logger.Info("validating client private key PEM block")
		if len(pemBlocks) > 1 {
			return fmt.Errorf("found more than one PEM block in client private key")
		}
		if pemBlocks[0].Type != pemBlockTypeRSAPrivateKey {
			return fmt.Errorf("found incorrect PEM block type in client private key")
		}
		cm.logger.Info("parsing client private key PEM block")
		privKey, err = x509.ParsePKCS1PrivateKey(pemBlocks[0].Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PEM bytes of client private key")
		}
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
	}

	cm.clientPrivKey = privKey

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

func (cm *certificateManager) GetMTLSClient() (*http.Client, error) {
	if cm.clientCert == nil {
		return nil, fmt.Errorf("cannot create mTLS client with nil client certificate")
	}
	if cm.clientPrivKey == nil {
		return nil, fmt.Errorf("cannot create mTLS client with nil client private key")
	}
	if cm.caCert == nil {
		return nil, fmt.Errorf("cannot create mTLS client with nil CA certificate")
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

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

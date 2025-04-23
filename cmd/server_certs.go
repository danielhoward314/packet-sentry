package cmd

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

type ServerCertBundle struct {
	ServerCert tls.Certificate
	CACert     *x509.Certificate
	CAKey      *rsa.PrivateKey
}

func LoadServerCerts(serverCertPath, serverKeyPath, caCertPath, caKeyPath string, includeCAKey bool) (*ServerCertBundle, error) {
	if serverCertPath == "" {
		return nil, fmt.Errorf("failed to load server cert path from env var")
	}
	if serverKeyPath == "" {
		return nil, fmt.Errorf("failed to load server key path from env var")
	}
	if caCertPath == "" {
		return nil, fmt.Errorf("failed to load CA cert path from env var")
	}
	if includeCAKey && caKeyPath == "" {
		return nil, fmt.Errorf("failed to load CA key path from env var")
	}

	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading server cert key pair: %w", err)
	}
	caCertPEMBytes, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("error reading CA cert: %w", err)
	}
	block, _ := pem.Decode(caCertPEMBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA cert: %v", err)
	}

	if !includeCAKey {
		return &ServerCertBundle{
			ServerCert: serverCert,
			CACert:     caCert,
		}, nil
	}

	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA key: %v", err)
	}
	block, _ = pem.Decode(caKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA key PEM")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %v", err)
	}
	caKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parsed key is not an RSA private key")
	}
	return &ServerCertBundle{
		ServerCert: serverCert,
		CACert:     caCert,
		CAKey:      caKey,
	}, nil
}

func LoadServerTLSCreds(certs *ServerCertBundle, isMTLS bool) credentials.TransportCredentials {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certs.ServerCert},
		MinVersion:   tls.VersionTLS12,
	}

	if isMTLS {
		certPool := x509.NewCertPool()
		certPool.AddCert(certs.CACert)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return credentials.NewTLS(tlsConfig)
}

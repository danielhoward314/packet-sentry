package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
	svcAgent "github.com/danielhoward314/packet-sentry/services/agent"
	svcBootstrap "github.com/danielhoward314/packet-sentry/services/bootstrap"
)

type certs struct {
	serverCert tls.Certificate
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	certs, err := loadCerts(
		"SERVER_CERT_PATH",
		"SERVER_KEY_PATH",
		"CA_CERT_PATH",
		"CA_KEY_PATH",
	)
	if err != nil {
		log.Fatalf("failed to load TLS creds")
	}

	tlsCreds := loadTLSCreds(certs, false)
	mtlsCreds := loadTLSCreds(certs, true)

	// Create gRPC servers
	tlsServer := grpc.NewServer(grpc.Creds(tlsCreds))
	mtlsServer := grpc.NewServer(grpc.Creds(mtlsCreds))

	bootstrapService := svcBootstrap.NewBootstrapService(
		logger,
		certs.caCert,
		certs.caKey,
	)
	pbBootstrap.RegisterBootstrapServiceServer(tlsServer, bootstrapService)

	agentService := svcAgent.NewAgentService(logger)
	pbAgent.RegisterAgentServiceServer(mtlsServer, agentService)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start TLS server
	wg.Add(1)
	go serveGRPC(ctx, &wg, tlsServer, ":9443", "TLS", logger)

	// Start mTLS server
	wg.Add(1)
	go serveGRPC(ctx, &wg, mtlsServer, ":9444", "mTLS", logger)

	// Wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	logger.Info("shutdown signal received")
	cancel()

	// Shutdown servers gracefully
	go shutdownGRPC(tlsServer, "TLS", logger)
	go shutdownGRPC(mtlsServer, "mTLS", logger)

	wg.Wait()
	logger.Info("all servers shut down cleanly")
}

func serveGRPC(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, addr string, label string, logger *slog.Logger) {
	defer wg.Done()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to bind", slog.String("label", label), slog.String("address", addr), slog.Any("error", err))
		return
	}
	logger.Info("starting server", slog.String("label", label), slog.String("address", addr))

	go func() {
		<-ctx.Done()
		_ = lis.Close() // triggers server.Serve to return
	}()

	if err := server.Serve(lis); err != nil && ctx.Err() == nil {
		logger.Error("server exited with error", slog.String("label", label), slog.Any("error", err))
	}
}

func shutdownGRPC(server *grpc.Server, label string, logger *slog.Logger) {
	logger.Info("initiating graceful shutdown", slog.String("label", label))

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("server shutdown completed", slog.String("label", label))
	case <-time.After(10 * time.Second):
		logger.Warn("shutdown timed out, forcing stop", slog.String("label", label))
		server.Stop()
	}
}

func loadCerts(serverCertEnvVar, keyEnvVar, caCertEnvVar, caKeyEnvVar string) (*certs, error) {
	serverCertPath := os.Getenv(serverCertEnvVar)
	serverKeyPath := os.Getenv(keyEnvVar)
	caCertPath := os.Getenv(caCertEnvVar)
	caKeyPath := os.Getenv(caKeyEnvVar)

	if serverCertPath == "" {
		return nil, fmt.Errorf("failed to load server cert path from env var")
	}
	if serverKeyPath == "" {
		return nil, fmt.Errorf("failed to load server key path from env var")
	}
	if caCertPath == "" {
		return nil, fmt.Errorf("failed to load CA cert path from env var")
	}
	if caKeyPath == "" {
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
		return nil, fmt.Errorf("Failed to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse CA cert: %v", err)
	}
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read CA key: %v", err)
	}
	block, _ = pem.Decode(caKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("Failed to decode CA key PEM")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse PKCS#8 private key: %v", err)
	}
	caKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("Parsed key is not an RSA private key")
	}
	return &certs{
		serverCert: serverCert,
		caCert:     caCert,
		caKey:      caKey,
	}, nil
}

func loadTLSCreds(certs *certs, isMTLS bool) credentials.TransportCredentials {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certs.serverCert},
		MinVersion:   tls.VersionTLS12,
	}

	if isMTLS {
		certPool := x509.NewCertPool()
		certPool.AddCert(certs.caCert)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return credentials.NewTLS(tlsConfig)
}

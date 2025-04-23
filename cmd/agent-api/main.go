package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/danielhoward314/packet-sentry/cmd"
	pbAgent "github.com/danielhoward314/packet-sentry/protogen/golang/agent"
	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
	"github.com/danielhoward314/packet-sentry/services"
)

const (
	serverCertPath = "certs/agent_server.cert.pem"
	serverKeyPath  = "certs/agent_server.key.pem"
	caCertPath     = "certs/ca.cert.pem"
	caKeyPath      = "certs/ca.key.pem"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	agentAPIPort := os.Getenv("AGENT_API_PORT")
	if agentAPIPort == "" {
		log.Fatalf("failed to get AGENT_API_PORT from env var")
	}
	agentAPIMTLSPort := os.Getenv("AGENT_API_MTLS_PORT")
	if agentAPIMTLSPort == "" {
		log.Fatalf("failed to get AGENT_API_MTLS_PORT from env var")
	}
	apiAddr := ":" + agentAPIPort
	apiMTLSAddr := ":" + agentAPIMTLSPort

	certs, err := cmd.LoadServerCerts(
		serverCertPath,
		serverKeyPath,
		caCertPath,
		caKeyPath,
		true,
	)
	if err != nil {
		log.Fatalf("failed to load TLS creds")
	}

	tlsCreds := cmd.LoadServerTLSCreds(certs, false)
	mtlsCreds := cmd.LoadServerTLSCreds(certs, true)

	// Create gRPC servers
	tlsServer := grpc.NewServer(grpc.Creds(tlsCreds))
	mtlsServer := grpc.NewServer(grpc.Creds(mtlsCreds))

	bootstrapService := services.NewBootstrapService(
		logger,
		certs.CACert,
		certs.CAKey,
	)
	pbBootstrap.RegisterBootstrapServiceServer(tlsServer, bootstrapService)

	agentService := services.NewAgentService(logger)
	pbAgent.RegisterAgentServiceServer(mtlsServer, agentService)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start TLS server
	wg.Add(1)
	go serveGRPC(ctx, &wg, tlsServer, apiAddr, "TLS", logger)

	// Start mTLS server
	wg.Add(1)
	go serveGRPC(ctx, &wg, mtlsServer, apiMTLSAddr, "mTLS", logger)

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

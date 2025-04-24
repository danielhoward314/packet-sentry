package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/gomail.v2"

	psPostgres "github.com/danielhoward314/packet-sentry/dao/postgres"
	psRedis "github.com/danielhoward314/packet-sentry/dao/redis"
	accountspb "github.com/danielhoward314/packet-sentry/protogen/golang/accounts"
	authpb "github.com/danielhoward314/packet-sentry/protogen/golang/auth"
	orgspb "github.com/danielhoward314/packet-sentry/protogen/golang/organizations"
	"github.com/danielhoward314/packet-sentry/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	user := os.Getenv("POSTGRES_USER")
	applicationDB := os.Getenv("POSTGRES_APPLICATION_DATABASE")
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		applicationDB,
		sslMode,
	)
	logger.Info("connecting to postgres")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the postgres:", err)
	}
	defer db.Close()

	logger.Info("connecting to redis")
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("error: REDIS_HOST is empty")
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		log.Fatal("error: REDIS_PORT is empty")
	}
	redisAddr := redisHost + ":" + redisPort
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // use default DB
	})

	logger.Info("connecting to SMTP server")
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		log.Fatal("error: SMTP_HOST is empty")
	}
	smtpPortStr := os.Getenv("SMTP_PORT")
	if smtpPortStr == "" {
		log.Fatal("error: SMTP_PORT is empty")
	}
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Fatal("error: invalid SMTP_PORT")
	}
	smtpDialer := gomail.NewDialer(smtpHost, smtpPort, "", "")
	smtpDialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// secret for JWT access token generation
	accessTokenJWTSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	if accessTokenJWTSecret == "" {
		log.Fatal("error: ACCESS_TOKEN_SECRET is empty")
	}

	// secret for JWT refresh token generation
	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	if refreshTokenSecret == "" {
		log.Fatal("error: REFRESH_TOKEN_SECRET is empty")
	}

	// secret for JWT install key generation
	installKeySecret := os.Getenv("INSTALL_KEY_SECRET")
	if installKeySecret == "" {
		log.Fatal("error: INSTALL_KEY_SECRET is empty")
	}

	datastore := psPostgres.NewDatastore(db, installKeySecret)
	registrationDatastore := psRedis.NewRegistrationDatastore(redisClient)
	tokenDatastore := psRedis.NewTokenDatastore(redisClient, accessTokenJWTSecret, refreshTokenSecret)

	logger.Info("injecting dependencies into service layer")
	// TODO: inject context into all of the services
	accountSvc := services.NewAccountsService(
		datastore,
		registrationDatastore,
		tokenDatastore,
		smtpDialer,
	)

	authSvc := services.NewAuthService(
		datastore,
		tokenDatastore,
		smtpDialer,
	)

	organizationsSvc := services.NewOrganizationsService(
		datastore,
	)

	accountspb.RegisterAccountsServiceServer(server, accountSvc)
	authpb.RegisterAuthServiceServer(server, authSvc)
	orgspb.RegisterOrganizationsServiceServer(server, organizationsSvc)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	apiPort := os.Getenv("API_PORT")
	if len(apiPort) == 0 {
		apiPort = "50051"
	}
	apiAddr := "[::]" + ":" + apiPort
	go serveGRPC(ctx, &wg, server, apiAddr, "web-api", logger)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	logger.Info("shutdown signal received")
	cancel()

	// Shutdown servers gracefully
	go shutdownGRPC(server, "web-api", logger)

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

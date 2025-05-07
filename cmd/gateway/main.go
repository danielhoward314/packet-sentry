package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/danielhoward314/packet-sentry/middleware"
	pbAccounts "github.com/danielhoward314/packet-sentry/protogen/golang/accounts"
	pbAdministrators "github.com/danielhoward314/packet-sentry/protogen/golang/administrators"
	pbAuth "github.com/danielhoward314/packet-sentry/protogen/golang/auth"
	pbDevices "github.com/danielhoward314/packet-sentry/protogen/golang/devices"
	pbOrgs "github.com/danielhoward314/packet-sentry/protogen/golang/organizations"
)

const (
	serverCertPath = "certs/gateway_server.cert.pem"
	serverKeyPath  = "certs/gateway_server.key.pem"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	apiHost := os.Getenv("API_HOST")
	if len(apiHost) == 0 {
		apiHost = "localhost"
	}
	apiPort := os.Getenv("API_PORT")
	if len(apiPort) == 0 {
		apiPort = "50051"
	}
	apiAddr := apiHost + ":" + apiPort
	// TODO: service mesh / envoy sidecar for mTLS between gateway and web-api
	conn, err := grpc.NewClient(apiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to hello service: %v", err)
	}
	defer conn.Close()

	logger.Info("setting up grpc-gateway serve mux")
	mux := runtime.NewServeMux()

	logger.Info("registering accounts service on gateway mux")
	err = pbAccounts.RegisterAccountsServiceHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register the accounts service handler: %v", err)
	}

	logger.Info("registering auth service on gateway mux")
	err = pbAuth.RegisterAuthServiceHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register the auth service handler: %v", err)
	}

	logger.Info("registering organizations service on gateway mux")
	err = pbOrgs.RegisterOrganizationsServiceHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register the organizations service handler: %v", err)
	}

	logger.Info("registering administrators service on gateway mux")
	err = pbAdministrators.RegisterAdministratorsServiceHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register the administrators service handler: %v", err)
	}

	logger.Info("registering devices service on gateway mux")
	err = pbDevices.RegisterDevicesServiceHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalf("failed to register the devices service handler: %v", err)
	}

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

	accessTokenJWTSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	if accessTokenJWTSecret == "" {
		log.Fatal("error: ACCESS_TOKEN_SECRET is empty")
	}

	// endpoints the authorization middleware skips checking API access token
	pathsWithoutAuthorization := []string{
		"/v1/signup",
		"/v1/verify",
		"/v1/login",
		"/v1/session",
		"/v1/refresh",
		"/v1/reset-verify",
		"/v1/passwords",
	}
	// lists of endpoints that only primary admins are authorized for
	// the authorization middelware uses the authorization_role claim
	// within the access token JWT
	primaryAdminEndpoints := []string{
		"/v1/install-keys",
	}

	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	authMiddleware := middleware.NewAuthMiddleware(redisClient, accessTokenJWTSecret, pathsWithoutAuthorization, primaryAdminEndpoints)

	middlewareWrappedMux := loggingMiddleware(authMiddleware(mux))

	corsEnv := os.Getenv("CORS_ALLOW_LIST")
	if len(corsEnv) == 0 {
		corsEnv = "http://localhost:5173"
	}
	corsAllowList := strings.Split(corsEnv, ",")
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   corsAllowList,                                       // Allow only these origins
		AllowedMethods:   []string{"OPTIONS", "GET", "POST", "PUT", "DELETE"}, // Allow specific methods
		AllowedHeaders:   []string{"Authorization", "Content-Type"},           // Allow specific headers
		AllowCredentials: true,                                                // Allow credentials
	}).Handler(middlewareWrappedMux)

	gatewayAddr := "[::]" + ":" + "8080"
	server := http.Server{
		Addr:         gatewayAddr,
		Handler:      corsHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		err := server.ListenAndServeTLS(serverCertPath, serverKeyPath)
		if err != nil {
			log.Fatalf("gateway server failed to listen: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	logger.Info("shutdown signal received", "signal", sig)

	ctx, cancelFunc := context.WithTimeout(ctx, 30*time.Second)
	defer cancelFunc()
	server.Shutdown(ctx)
}

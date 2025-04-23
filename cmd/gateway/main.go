package main

import (
	"context"
	"log"
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
	accountspb "github.com/danielhoward314/packet-sentry/protogen/golang/accounts"
	authpb "github.com/danielhoward314/packet-sentry/protogen/golang/auth"
	orgspb "github.com/danielhoward314/packet-sentry/protogen/golang/organizations"
)

func main() {
	apiHost := os.Getenv("API_HOST")
	if len(apiHost) == 0 {
		apiHost = "localhost"
	}
	apiPort := os.Getenv("API_PORT")
	if len(apiPort) == 0 {
		apiPort = "50051"
	}
	apiAddr := apiHost + ":" + apiPort
	conn, err := grpc.NewClient(apiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to hello service: %v", err)
	}
	defer conn.Close()
	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	err = accountspb.RegisterAccountsServiceHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalf("failed to register the accounts service handler: %v", err)
	}
	err = authpb.RegisterAuthServiceHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalf("failed to register the auth service handler: %v", err)
	}
	err = orgspb.RegisterOrganizationsServiceHandler(context.Background(), mux, conn)
	if err != nil {
		log.Fatalf("failed to register the organizations service handler: %v", err)
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
	}
	// lists of endpoints that only primary admins are authorized for
	// the authorization middelware uses the authorization_role claim
	// within the access token JWT
	primaryAdminEndpoints := []string{
		"/v1/organizations", // TODO: remove this route, only here to demonstrate the auth middleware works
	}

	authMiddleware := middleware.NewAuthMiddleware(redisClient, accessTokenJWTSecret, pathsWithoutAuthorization, primaryAdminEndpoints)

	middlewareWrappedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMiddleware.ServeHTTP(w, r, mux)
	})

	corsEnv := os.Getenv("CORS_ALLOW_LIST")
	if len(corsEnv) == 0 {
		corsEnv = "http://localhost:5173"
	}
	corsAllowList := strings.Split(corsEnv, ",")
	// Set up CORS middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   corsAllowList,                                       // Allow only this origin
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

	// start server in its own goroutine
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("gateway server failed to listen: %v", err)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// block until a signal is received
	sig := <-sigs
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()
	server.Shutdown(ctx)
}

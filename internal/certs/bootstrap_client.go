package certs

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pbBootstrap "github.com/danielhoward314/packet-sentry/protogen/golang/bootstrap"
)

// NewBootstrapGRPCClient creates a BootstrapServiceClient with TLS in prod or insecure in dev
func NewBootstrapGRPCClient(ctx context.Context, addr string, isDev bool) (pbBootstrap.BootstrapServiceClient, error) {
	var dialOpts grpc.DialOption

	// TODO: get rid of this bool and instead document putting the self-signed CA cert in trusted system certs
	if isDev {
		// In dev, skip cert verification
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // ideally this is `false` and local dev trusts self-signed CA cert so we can still do this
			MinVersion:         tls.VersionTLS12,
			ServerName:         "localhost",
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = grpc.WithTransportCredentials(creds)

	} else {
		// In prod, use system-trusted TLS
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			ServerName:         "packet-sentry.com",
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = grpc.WithTransportCredentials(creds)
	}

	conn, err := grpc.NewClient(addr, dialOpts)
	if err != nil {
		return nil, err
	}

	return pbBootstrap.NewBootstrapServiceClient(conn), nil
}

# Repo Structure

The `cmd` directory contains several Go binaries:

1. The agent `cmd/agent` is the binary that gets installed on Linux, macOS, and Windows endpoints to send network telemetry.
2. The agent-api `cmd/agent-api` is the gRPC server with which the agent communicates to receive configuration and to send its telemetry.
3. The web-api `cmd/web-api` is the gRPC server code that handles the business logic of the API that powers the web console.
4. The gateway `cmd/gateway` uses the Google grpc-gateway to translate JSON RESTful API requests to protobufs and reverse proxies them to the web-api.
5. The cli `cmd/cli` is a CLI tool for managing the application database and SQL migrations for its tables.
6. The installer actions `cmd/installeractions` are used by the WiX-based MSI as custom actions for the Windows agent installer.


The `packet-sentry-web-console` directory contains the React SPA for the Packet Sentry Web Console.
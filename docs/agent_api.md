# agent-api

This Go executable runs two gRPC servers. One is for bootstrapping trust between the agent and the agent-api, using TLS communication. The other uses mutual TLS.

Run it (TODO: refine these instructions once env-based config is set up)

```
SERVER_CERT_PATH=/home/danielhoward/repos/scratch/cmd/agent-api/server.cert.pem SERVER_KEY_PATH=/home/danielhoward/repos/scratch/cmd/agent-api/server.key.pem CA_CERT_PATH=/home/danielhoward/repos/scratch/cmd/agent-api/ca.cert.pem CA_KEY_PATH=/home/danielhoward/repos/scratch/cmd/agent-api/ca.key.pem go run cmd/agent-api/main.go
```
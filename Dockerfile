FROM golang:1.24 AS gobase

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy all .go files
COPY . .

COPY /certs /certs

COPY /templates /templates

# Build the agent-api binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /packet-sentry-agent-api ./cmd/agent-api/main.go

# Build the cli binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /packet-sentry-cli ./cmd/cli/main.go

# Build the gateway binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /packet-sentry-gateway ./cmd/gateway/main.go

# Build the web-api binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /packet-sentry-web-api ./cmd/web-api/main.go

############################################
# agent-api
############################################
FROM scratch AS agent-api
COPY --from=gobase /packet-sentry-agent-api /bin/agent-api
COPY --from=gobase /certs/ca.cert.pem /certs/ca.cert.pem
COPY --from=gobase /certs/ca.key.pem /certs/ca.key.pem
COPY --from=gobase /certs/agent_server.cert.pem /certs/agent_server.cert.pem
COPY --from=gobase /certs/agent_server.key.pem /certs/agent_server.key.pem
EXPOSE 9443
EXPOSE 9444
CMD ["/bin/agent-api"]

############################################
# cli
############################################
FROM ubuntu:latest AS cli
COPY --from=gobase /packet-sentry-cli /bin/cli
RUN chmod +x /bin/cli
ENTRYPOINT ["/bin/cli"]

############################################
# gateway
############################################
FROM scratch AS gateway
COPY --from=gobase /packet-sentry-gateway /bin/gateway
COPY --from=gobase /certs/ca.cert.pem /certs/ca.cert.pem
COPY --from=gobase /certs/gateway_server.cert.pem /certs/gateway_server.cert.pem
COPY --from=gobase /certs/gateway_server.key.pem /certs/gateway_server.key.pem
EXPOSE 8080
CMD ["/bin/gateway"]

############################################
# web-api
############################################
FROM scratch AS web-api
COPY --from=gobase /packet-sentry-web-api /bin/web-api
COPY --from=gobase /certs/ca.cert.pem /certs/ca.cert.pem
COPY --from=gobase /certs/web_api_server.cert.pem /certs/web_api_server.cert.pem
COPY --from=gobase /certs/web_api_server.key.pem /certs/web_api_server.key.pem
COPY --from=gobase /templates /templates
EXPOSE 50051
CMD ["/bin/web-api"]

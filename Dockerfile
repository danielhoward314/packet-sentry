FROM golang:1.24 as gobase

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy all .go files
COPY . .

COPY /certs /certs

COPY /templates /templates

# Build the agent-api binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /packet-sentry-agent-api ./cmd/agent-api/main.go

# Build the cli binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /packet-sentry-cli ./cmd/cli/main.go

# Build the gateway binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /packet-sentry-gateway ./cmd/gateway/main.go

# Build the web-api binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /packet-sentry-web-api ./cmd/web-api/main.go

############################################
# agent-api
############################################
FROM scratch as agent-api
COPY --from=gobase /packet-sentry-agent-api /bin/agent-api
COPY --from=gobase /certs /certs
EXPOSE 9443
EXPOSE 9444
CMD ["/bin/agent-api"]

############################################
# cli
############################################
FROM ubuntu:latest as cli
COPY --from=gobase /packet-sentry-cli /bin/cli
RUN chmod +x /bin/cli
ENTRYPOINT ["/bin/cli"]

############################################
# gateway
############################################
FROM scratch as gateway
COPY --from=gobase /packet-sentry-gateway /bin/gateway
EXPOSE 8080
CMD ["/bin/gateway"]

############################################
# web-api
############################################
FROM scratch as web-api
COPY --from=gobase /packet-sentry-web-api /bin/web-api
COPY --from=gobase /templates /templates
EXPOSE 50051
CMD ["/bin/web-api"]

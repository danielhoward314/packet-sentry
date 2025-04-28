# NATS

NATS has a concept of client applications and NATS service infrastructure. Quoting the [NATS documentation](https://docs.nats.io/nats-concepts/what-is-nats):

> Developers use one of the NATS client libraries in their application code to allow them to publish, subscribe, request and reply between instances of the application or between completely separate applications.

> The NATS services are provided by one or more NATS server processes that are configured to interconnect with each other and provide a NATS service infrastructure. The NATS service infrastructure can scale from a single NATS server process running on an end device (the nats-server process is less than 20 MB in size!) all the way to a public global super-cluster of many clusters spanning all major cloud providers and all regions of the world.

## nats-server

The `nats-server` container is defined in the `compose.yml` file. The `agent-api` and `web-api` containers depend on it, so it gets spun up when those containers get started.

## nats clients

The `agent-api` and `web-api` use the NATS Go client from package `github.com/nats-io/nats.go`. Jet Stream is used with streams for commands and packet events. The subjects are `cmds.*` and `packetEvents.*` where the wildcard is the same unique OS identifier used as the common name in the client certificate each agent uses for mTLS with the agent-api. The two main use cases are for commands and packet events.

1. The agent uses a unary gRPC client to poll the agent-api for commands intended for this device. Different parts of the backend publish commands for specific devices, which the agent-api will send to the agent as the agent polls.
2. The agent uses a streaming gRPC to send packet capture events. The agent-api gRPC server handler will receive these event streams from each device and publish them to NATS. The data platform subscribes to these events to prepare them for dashboards and telemetry insights in the web-console.
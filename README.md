# Packet Sentry

I created Packet Sentry to have a fully-featured SaaS product in my portfolio of personal projects.

Packet Sentry is a Network Detection and Response (NDR) platform with an agent deployed on endpoints and a web console with dashboards for monitoring agent deployments and the telemetry sent up from endpoints.

For more details, read my dev log [here](https://danielhoward-dev.netlify.app/).

## Running with docker compose

A prerequisite is to modify your hosts file (Unix `/etc/hosts`, Windows `C:\Windows\System32\drivers\etc\hosts`) so some hostnames will resolve to the loopback address:

```
127.0.0.1 web-console.packet-sentry.local
::1 web-console.packet-sentry.local
127.0.0.1 gateway.packet-sentry.local
::1 gateway.packet-sentry.local
127.0.0.1 agent-api.packet-sentry.local
::1 agent-api.packet-sentry.local
```

Create a config file (Unix `/opt/packet-sentry/config.json`, Windows `C:\Program Files\PacketSentry\config.json`) that will override the hostname and ports the agent uses for requests to the agent-api in local dev:

```
{"agentHost":"agent-api.packet-sentry.local", "agentPort":"9444", "bootstrapPort":"9443"}
```

Generate the local certs (TODO: write a PowerShell script for it as well):

```
./scripts/generate_certs
```

Use docker compose to build the `agent-api`, `cli`, `gateway`, `web-api`, and `web-console` containers:

```
docker compose --file compose.yml build
```

Initialize the database:

```
docker compose up -d postgres
docker compose run --rm cli create db
docker compose run --rm cli migrate up
```

Run the built containers:

```
docker compose --file compose.yml up agent-api gateway web-api web-console
```

The `web-console` container is an nginx server that serves the vite production build of the Vue SPA as static files.

In order to create a Packet Sentry account, navigate to `web-console.packet-sentry.local/signup`. For return visits, log in at `web-console.packet-sentry.local/login`.

## Documentation

- [Architecture](./docs/architecture.md)
- [Repo Structure](./docs/repo_structure.md)
- [Agent](./docs/agent.md)
- Agent Installers: [Linux](./docs/agent_installer_linux.md), [macOS](./docs/agent_installer_macos.md), and [Windows](./docs/agent_installer_windows)
- [DB Migrations](./docs/db_migrations.md)
- [Protos](./docs/protos.md)
- [NATS](./docs/nats.md)
- [authorization](./docs/authorization.md)
- [Local containers](./docs/local_containers.md) such as `redis` or `maildev`
- [tailwindcss](./docs/tailwindcss.md)
- [Vue component hierarchy](./docs/vue_component_hierarchy.md)
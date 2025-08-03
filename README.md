# Packet Sentry

I created Packet Sentry to have a fully-featured SaaS product in my portfolio of personal projects.

Packet Sentry is a Network Detection and Response (NDR) platform with an agent deployed on endpoints and a web console. The agent is cross-platform (Windows, macOS, and Linux), runs as a daemon on Unix and as a background Service on Windows. The web console is a single pane of glass for network telemetry across all installations.

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

Generate the local certs on Unix:

```bash
./scripts/generate_certs
```

Or on Windows:

```powershell
.\scripts\generate_certs.ps1 # assumes openssl is installed and in PATH
```

You have to trust the self-signed root CA at `./certs/ca.cert.pem`.

Use docker compose to build the `agent-api`, `cli`, `gateway`, `web-api`, and `web-console` containers:

```
docker compose --file compose.yml build
```

Initialize the database:

```
docker compose up -d postgres
docker compose run --rm cli create db
docker compose run --rm cli migrate up
docker compose run --rm cli migrate up --timescale
```

Run the built containers:

```
docker compose --file compose.yml up agent-api gateway web-api web-console worker
```

The `web-console` container is an nginx server that serves the vite production build of the React SPA as static files.

In order to create a Packet Sentry account, navigate to `web-console.packet-sentry.local/signup`. For return visits, log in at `web-console.packet-sentry.local/login`.

## Documentation

- [Repo Structure](./docs/repo_structure.md)
- [Agent](./docs/agent.md)
- Agent Installers: [Linux](./docs/agent_installer_linux.md), [macOS](./docs/agent_installer_macos.md), and [Windows](./docs/agent_installer_windows)
- [DB Migrations](./docs/db_migrations.md)
- [Protos](./docs/protos.md)
- [NATS](./docs/nats.md)
- [authorization](./docs/authorization.md)
- [Local containers](./docs/local_containers.md) such as `redis` or `maildev`
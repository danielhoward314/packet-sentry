## DB Migrations

The `cli` container is a [cobra CLI](https://github.com/spf13/cobra) that exposes commands for initializing the application database. The CLI exposes additional commands that leverage [goose](https://github.com/pressly/goose) for executing the SQL migration files against the application database.

Build the `cli`:

```
docker compose build cli
```

Run the `postgres` container:

```
docker compose up -d postgres
```

To create the main application database:

```
docker compose run --rm cli create db
```

This will create a database with the name matching the `POSTGRES_APPLICATION_DATABASE` value in `env/postgres`.

To drop the main application database:

```
docker compose run --rm cli drop db
```

To create a SQL migration file for the main application database or the TimescaleDB database:

```
docker compose run --rm cli create migration <name>
docker compose run --rm cli create migration <name> --timescale
```

This will create an empty migration file of the format `<timestamp>_<name>.sql` where `<name>` corresponds to the given CLI argument. The file will be generated in the `cmd/cli/commands/migrations` directory.

To run all migrations through goose:

```
docker compose run --rm cli migrate up
docker compose run --rm cli migrate up --timescale
```

The `goose` tool maintains its own table tracking which migrations have been run. The `up` command will only run migrations that have yet to be run, allowing the lifecycle of the database schemas to exist in version control.

To roll back all migrations through goose:

```
docker compose run --rm cli migrate down
docker compose run --rm cli migrate down --timescale
```

For troubleshooting the `postgres` container, you can exec into it with this command:

```
docker exec -it postgres psql -U postgres postgres
\c packet_sentry
\c packet_sentry_timescale
```

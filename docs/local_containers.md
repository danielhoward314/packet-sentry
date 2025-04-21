## Local Containers

### redis

The `redis` container is used during the signup process for email verification data and is used for user session JWTs. For troubleshooting the `redis` container, you can exec into it and run the Redis CLI with these commands:

```bash
docker exec -it redis sh
redis-cli
```

Some common commands are:

```bash
KEYS * # get all keys by a pattern (wildcard used here)
GET <key> # read the data at a given key
FLUSHDB # delete all data
```

### maildev SMTP mock server

The signup flow for this application requires new users verify their email with a code sent via email. In local development, I use [maildev](https://github.com/maildev/maildev) as a mock SMTP server. In addition to the SMPT server, the `maildev` container also spins up a UI at `http://0.0.0.0:1080/`.
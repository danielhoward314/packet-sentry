# Testing the Web API

The `gateway` service exposes a RESTful API that reverse proxies requests to the `web-api`.

## Pre-requisites

The containers depend on their server certs:

```bash
./scripts/generate_certs
```

Build and run both required containers (the `compose.yml` will spin up the containers they depend on):

```bash
docker compose --file compose.yml build
docker compose --file compose.yml up web-api gateway
```

## Test endpoints

### POST v1/signup

To confirm that the container is reachable from your host:

```bash
curl gateway.packet-sentry.local:8080/v1/signup
```

If you get an error about not being able to resolve the host, you need to modify your hosts file (see the main README.md). Otherwise, you should get a response like `{"code":12,"message":"Method Not Allowed","details":[]}`.

Make a signup request with all required parameters:

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/signup \
    -H "Content-Type: application/json" \
    -d '{"organizationName": "testname", "primaryAdministratorEmail": "testadmin@testorg.com", "primaryAdministratorName": "testadminname", "primaryAdministratorCleartextPassword": "testpassword"}'
```

Happy path side-effects:

1. Should send JSON response with a token `{"token":"<token>"}`.
2. The service should have persisted in Redis at a key matching `<token>` a JSON string with the email verification code and other Postgres-persisted data for the new organization.

```bash
docker exec -it redis sh
redis-cli
KEYS * # show all keys
GET <token>
"{\"organization_id\":\"<org-uuid>\",\"administrator_id\":\"<admin-uuid>\",\"email_code\":\"<email-code>\"}"
```

3. The service should have persisted in Postgres the newly created organization and administrator.

```bash
docker exec -it postgres psql -U postgres postgres
\c packet_sentry
select * from organizations;
select * from administrators;
```

Notably, the `administrators` row will have in the `password_hash` column a bcrypt hash of the cleartext password from the request and will have its `verified` column value set to `false`. The web console expects users to validate their email with a verification code; this happens with successful calls to the `/v1/verify` endpoint.

4. The SMTP server's UI (`maildev` container spins this up at `http://0.0.0.0:1080/`) should have an email in the mock-inbox with the verification code matching the `<email-code>` from the Redis JSON data. This is mocking out the typical signup flow that requires email verification.
5. Since the Redis TTL is 10 minutes, you should get no data back from the redis-cli after this TTL elapses.

Make the same request again and it should fail due to the uniqueness constraint on the primary administrator email.

Make the same request with any missing parameters and it should fail.

### POST v1/verify

This endpoint requires the token and email code from the `/v1/signup` endpoint.

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/verify \
    -H "Content-Type: application/json" \
    -d '{"token": "<token>", "verificationCode": "<email-code>"}'
```

Happy path side-effects:

1. Should send JSON response with the access and refresh tokens:

```json
{
    "adminUiAccessToken":"<ui-access-token>",
    "adminUiRefreshToken":"<ui-refresh-token>",
    "apiAccessToken":"<api-access-token>",
    "apiRefreshToken":"<api-refresh-token>"
}
```

2. Should delete the Redis data for the token and verification code that was created as a side-effect of the `/v1/signup` endpoint.

3. Should persist in Redis 4 JWTs for the access and refresh tokens used for session management in the web console and for authorization of API calls. Below is an example of looking up the Redis data for this JWT (same applies for the other token types above):

```bash
docker exec -it redis sh
redis-cli

get <ui-access-token>
"{\"organization_id\":\"<organizations.id>\",\"administrator_id\":\"<administrators.id>\",\"authorization_role\":\"PRIMARY_ADMIN\",\"token_type\":<1|2>,\"claims_type\":<1|2>}"

ttl <ui-access-token>
(integer) 10785
```

The organization and administrator ids correspond to the primary keys for the rows in Postgres. The authorization role is the one granted to the admin who creates the account. The token types are 1 for access and 2 for refresh. The claims types are 1 for admin UI and 2 for API. The TTL is a number of seconds until expiry.

Use the [jwt debugger](https://jwt.io/) to see the decoded JWT. Note that for access tokens the expiry (`exp`) will be sooner in the JWT than in the corresponding access token TTL in Redis. This is to allow for detecting an expired, but valid, access token which is useful for signaling with error codes that the caller should use their refresh token to request a new access token. For refresh tokens, the JWT `exp` and Redis TTL are the same.

4. Should set the administrator's `verified` column value to `true`.

### POST /v1/session

This endpoint validates the admin UI session JWTs. The web console makes use of this endpoint to handle session management. When a user's admin UI sesssion

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/session \
    -H "Content-Type: application/json" \
    -d '{"jwt": "<ui-access-token>"}'
```

Assuming a valid JWT, this API should respond with the same JWT.

### POST /v1/refresh

This endpoint is used to request a new access token by providing a refresh token and the claims type. The claims type should correspond to the one in the refresh token, i.e. use claims type 1 for an admin UI refresh token and 2 for an API refresh token.

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/refresh \
    -H "Content-Type: application/json" \
    -d '{"jwt": "<ui|api-refresh-token>", "claimsType": <1|2>}'
```

Happy path side-effects:

1. A new access token is persisted in Redis for the claims type in the request.
2. The API responds with this new JWT.

### POST /v1/login

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/login \
    -H "Content-Type: application/json" \
    -d '{"email": "<email>", "password": "<password>"}'
```

This API should respond with the organization, administrator, and token data that the web console uses for context and session management.

### GET /v1/organizations/{id}

```bash
curl --cacert ./certs/ca.cert.pem -X GET https://gateway.packet-sentry.local:8080/v1/organizations/<organization-id> \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>"
```

This API should return the organization data.

### POST /v1/install-keys

```bash
curl --cacert ./certs/ca.cert.pem -X POST https://gateway.packet-sentry.local:8080/v1/install-keys \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>" \
    -d '{"administratorEmail": "<email>"}'
```

This API should persist an install key associated with this administrator and respond with the key.

### GET /v1/devices/{id}

```bash
curl --cacert ./certs/ca.cert.pem -X GET https://gateway.packet-sentry.local:8080/v1/devices/<device-id> \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>"
```

### GET /v1/devices

```bash
curl --cacert ./certs/ca.cert.pem -X GET https://gateway.packet-sentry.local:8080/v1/devices?organizationId=<org-id> \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>"
```

### PUT /v1/devices/{id}

```bash
curl --cacert ./certs/ca.cert.pem -X PUT https://gateway.packet-sentry.local:8080/v1/devices/750baff0-8c7f-4982-a0c8-04e415adfdae \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>" \
    -d '{"pcapVersion": "<version>", "clientCertPem": "<cert-pem>", "clientCertFingerprint": "<fingerprint>", "interfaces": ["<interface-name>"], "interface_bpf_associations": {"lo": {"captures": {"tcp port 3000": {"bpf": "tcp port 3000", "deviceName": "lo", "snaplen": 65535}}}}}'
```

### GET /v1/events/{deviceId}

```bash
curl --cacert ./certs/ca.cert.pem -X GET "https://gateway.packet-sentry.local:8080/v1/events/<device-id>?start=2025-05-26T01:00:00.000Z&end=2025-05-26T03:02:00.000Z" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <api-access-token>"
```
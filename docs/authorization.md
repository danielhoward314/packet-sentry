## Authorization

The backend of Packet Sentry uses JSON Web Tokens (JWT) for authorization. The JWT design uses access tokens and refresh tokens. A short-lived access token is used to authorize access to resources and when this expires, a longer-lived refresh token is used to request a new access token. Two pairs of access and refresh tokens exist: one for session management of an administrator's UI session and another for the API (XHR calls from the frontend or any API client, such as `curl`). The frontend sets these tokens in localStorage upon successful login. With the pair of tokens meant for admin UI sessions, the frontend guards client-side routes. With the pair of tokens meant for API authorization, the frontend uses the access token in XHR requests. The `gateway` container defines middleware that validates an access token JWT submitted in API requests via header as follows: `Authorization: Bearer <api_access_token>`.

### Role-based access control (RBAC)

In addition to checking the validity of the signature of the JWT, the application uses a custom claim in the JWT that designates the authorization role of the subject. The value of this claim corresponds to the `authorization_role` column of the `administrator`. Any administrator who completes the account signup is a primary admin; all others are created/modified with roles given by primary admins, defaulting to the secondary admin role. The authorization role claim is checked against the administrator's data to enforce role-based access control (RBAC) of resources.

### Axios Client

For XHR requests from the Vue SPA to the API, this application uses [axios](https://axios-http.com/docs/intro). A handful of routes for account creation and authentication use the native browser `fetch`, since these do not need the same interceptors. The axios client is used for the rest of the XHR calls.

The code that defines the base client configuration lives in `./packet-sentry-web-console/src/lib/axiosBaseClient.ts`. The file `./packet-sentry-web-console/src/lib/api.ts` defines the API calls that use the axios base client.

The base client defines an interceptor that checks for the presence of an API access token in localStorage and, when present, uses the token in the `Authorization: Bearer <api_access_token>` header. If the access token is invalid, the `gateway` container's middleware returns a 401 to allow API clients to use their refresh token to get a new access token. The interceptor implements this refresh call and, when successful, retries the original request.
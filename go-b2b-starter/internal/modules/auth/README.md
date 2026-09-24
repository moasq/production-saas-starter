# Authentication and tenant boundaries

Stytch B2B verifies each session with its `Sessions.Authenticate` API. The application reads the provider RBAC policy into a bounded five-minute process cache; permission failures deny access. There is no mock authentication or unsigned JWT mode. Empty credentials allow the setup page and API health check; signup returns 503 and protected requests remain unauthorized.

Configure Stytch custom role IDs `admin`, `manager`, and `member`. Add resource `org` with actions `view` and `manage`. Grant both actions to `admin` and `view` to the other roles. Signup assigns `admin`. These catalog definitions describe the recommended policy; they do not override provider permissions. A change to provider permissions can take up to five minutes to clear a process cache.

The server resolves the verified Stytch organization and member into organization-scoped PostgreSQL records. Every tenant operation uses this context, never an organization ID supplied in JSON. Organization and billing writes require `org:manage`. Profile updates use the session member and email.

Server callers pass an Authorization bearer token per request. Browsers use the HttpOnly `stytch_session_jwt` cookie. Cookie-authenticated mutations require an Origin matching `APP_BASE_URL` (or request Host when no URL is configured). Bearer calls do not use ambient browser credentials.

Set `STYTCH_PROJECT_ID`, `STYTCH_SECRET`, `STYTCH_ENV=test|live`, and both `STYTCH_LOGIN_REDIRECT_URL` and `STYTCH_INVITE_REDIRECT_URL`. Redirects normally point to your frontend `/authenticate` route. Partial credentials and mismatched environments are configuration errors. Provider secrets remain server-side.

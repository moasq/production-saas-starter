# Setup and deployment

## Run locally

Install Docker with Compose on Linux, macOS, or WSL, then run `./setup.sh` from the repository
root. It creates `.env` only if absent, generates a database password, builds the
app, waits for startup, and serves it at http://localhost:3000. A subsequent run
preserves your configuration and data. `docker compose down` stops services and
keeps volumes; do not add `--volumes` unless you intend to erase local data.

To configure manually, copy `.env.example` to `.env`, set a strong
`POSTGRES_PASSWORD`, and run `docker compose up --build -d --wait`.
Compose uses the checkout directory name as its project name. To run a second
checkout, set a distinct `COMPOSE_PROJECT_NAME`, `HTTP_PORT`, and `HTTPS_PORT`.
Use the same project name for subsequent start, logs, backup, and stop commands.

Only the reverse proxy publishes ports. PostgreSQL and Go remain on the private
Compose network. The initial build downloads Go/Node dependencies and can take
several minutes.

## Enable sign-in

Create a Stytch **B2B** test project and set these server-only values in `.env`:

```dotenv
STYTCH_ENV=test
STYTCH_PROJECT_ID=project-test-...
STYTCH_SECRET=your-project-secret
APP_BASE_URL=http://localhost:3000
```

In Stytch, allow `http://localhost:3000/authenticate` as a discovery and login
magic-link redirect. Use the same callback with your HTTPS domain in production.
The backend and frontend must use the same project and environment. Production
Stytch uses `STYTCH_ENV=live`; application build mode does not select a provider.
Before creating your first organization, add a resource `org` with actions `view`
and `manage` to the Stytch RBAC policy. Create custom roles `admin`, `manager`,
and `member`: admin receives both actions; manager and member receive `view`.
The starter assigns `admin` to the first organization member. These exact IDs are
part of the current API contract; provider roles remain authoritative.
See [backend authentication](go-b2b-starter/internal/modules/auth/README.md)
for the role/permission contract. Run `docker compose up -d` after changing `.env`.

Credentials are runtime variables. They are not Docker build arguments and there
are no browser-side Stytch tokens to configure. Missing credentials expose setup
instructions and fail authentication closed. Partially supplied or invalid
provider configuration should be corrected before real use.

## Enable optional billing

Create a recurring Polar product, then set:

```dotenv
BILLING_ENABLED=true
POLAR_ENVIRONMENT=sandbox
POLAR_ACCESS_TOKEN=your-organization-access-token
POLAR_PRODUCT_ID=your-recurring-product-id
```

Use one Polar environment consistently across both applications. Core pages do
not require billing. Organization administrators can start checkout and open the
customer portal. Subscription status comes from Polar's customer-state API on
request; this starter does not require a webhook endpoint or copy subscriptions
into PostgreSQL. Provider downtime is reported as an error. For custom product
entitlements or asynchronous fulfillment, add an explicit tested policy and,
if necessary, a durable webhook consumer.

Before taking real payments, use `POLAR_ENVIRONMENT=production` with production
credentials and verify a real checkout, portal, cancellation and renewal. A local
build or mocked test does not verify these external flows.

## Deploy on one server

Use a Linux host with Docker and Compose. Point a public DNS name to it and allow
inbound ports 80 and 443. Configure `.env` before starting:

```dotenv
APP_BASE_URL=https://app.example.com
SITE_ADDRESS=app.example.com
BIND_ADDRESS=0.0.0.0
HTTP_PORT=80
HTTPS_PORT=443
APP_ENV=PROD
```

Set real Stytch credentials and allowed callbacks for that domain. Run
`docker compose up --build -d --wait`. Caddy terminates HTTPS and renews
certificates; its data volume must persist. The frontend uses runtime configuration,
so the same image works with a different domain or provider credentials.
Container builds support the host's native architecture; there is no forced amd64
binary. Do not expose the API directly: the supplied deployment expects Caddy in
front of it. Proxy-to-application and database traffic stay on the private network.

Check `/api/health` (frontend) and `/health` (API readiness). Inspect
`docker compose ps` and `docker compose logs --tail=100` when startup fails.

## Backups and updates

Before an upgrade, back up PostgreSQL and test restoring into a separate database:

```sh
docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > starter.dump
```

The dump contains account data; store it privately. After pulling reviewed changes,
run `docker compose up --build -d --wait`. Image versions/digests and dependency
lockfiles are pinned; Dependabot proposes updates. Database major upgrades and
legacy installations need the [upgrade guide](docs/UPGRADING.md).

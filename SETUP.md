# Setup and deployment

## Run locally

Install Docker with Compose on Linux, macOS, or WSL, then run `./setup.sh` from
the repository root. It generates private database and auth secrets, builds the
app, applies migrations, and serves http://localhost:3000. Create a workspace and
open its sign-in link in the local inbox at http://localhost:8025. No Stytch or
Better Auth Cloud account is required.

A repeat run preserves configuration and data. `docker compose down` stops the
services and keeps volumes; adding `--volumes` erases local data. For a second
checkout, set a distinct `COMPOSE_PROJECT_NAME`, `HTTP_PORT`, `HTTPS_PORT`, and
`MAILPIT_PORT`. Set `APP_BASE_URL` to the matching URL. Use that project name for
subsequent start, logs, backup and stop commands.

To configure manually, copy `.env.example` to `.env`, generate independent random
values of at least 32 bytes for all five blank secrets, and run
`docker compose up --build -d --wait`. Keep `.env` private and out of source
control. The initial build downloads Go and Node dependencies.

Only Caddy and the loopback-only local mail inbox publish ports. PostgreSQL, Go,
and Next.js remain on the private Compose network. `database-init`,
`backend-migrate`, and `auth-migrate` are one-shot jobs: a successful exited job
is expected. An unsuccessful migration prevents startup.

## Diagnose setup

Run `./setup.sh --doctor` from any working directory by using the path to this
checkout's `setup.sh`. It checks Docker/Compose, configuration names and required
secret lengths, local port conflicts, service health and configuration drift, completed migration jobs,
the business schema version, auth migration checksums, database privileges, RLS,
and the allowed Better Auth membership roles. It does not need local Go or Node.
Use the same `COMPOSE_PROJECT_NAME` and environment overrides as setup.

The doctor reads configuration and performs read-only SQL. It never starts or
stops services, runs migrations, changes credentials, sends email, or calls Polar.
Values and raw Docker/database errors are suppressed. A missing `.env` or an
unstarted stack produces a failure with the next step; run `./setup.sh` to apply
setup, then rerun the doctor. Each Docker read has a 20-second deadline followed
by a one-second termination grace period. Set `DOCTOR_TIMEOUT_SECONDS` (1–120)
in the shell when a slow daemon needs a different deadline.

Exit status is `0` when these diagnostics pass, `1` for a detected problem, and
`2` when a check cannot be completed. Install `lsof` or `iproute2` (`ss`) to inspect
ports before services start, and `sha256sum` or `shasum` for migration checksums.
Remote Docker contexts need port verification on the remote host. A passing
doctor does not prove browser journeys, external email delivery or live billing.

## Windows support

Use **WSL2 with Docker Desktop's Linux-container integration**. Native PowerShell,
Command Prompt and Git Bash are not supported setup environments. Install Ubuntu
with `wsl --install -d Ubuntu`, confirm version 2 with `wsl --list --verbose`, and
enable that distribution under Docker Desktop **Settings → Resources → WSL
Integration**. Clone into the Linux home directory and run these in Ubuntu:

```sh
docker info
docker compose version
./setup.sh
./setup.sh --doctor
```

Keep shell scripts with LF endings. Follow the official
[WSL installation guide](https://learn.microsoft.com/en-us/windows/wsl/install)
and [Docker Desktop WSL2 guide](https://docs.docker.com/desktop/features/wsl/).
Automated shell tests run on Linux/macOS and exercise the Linux/WSL host contract;
the Docker deployment check runs on Linux. These checks do not emulate the Windows
Desktop integration or verify a user's WSL2 installation.

## Authentication and email

Better Auth stores sessions and organization membership in PostgreSQL. The
application has explicit admin, manager and member roles. Admin manages the
workspace and team; manager and member have workspace view access and can edit
their own profile. Users with multiple memberships select an active workspace.
Sessions have a fixed eight-hour lifetime, with no silent sliding renewal; users
request a fresh link afterward. Login links expire after ten minutes and are
single-use. Invitations expire after 48 hours. Existing Stytch links and
sessions do not transfer. See [Upgrading](docs/UPGRADING.md) for migration.

Development uses Mailpit (`COMPOSE_PROFILES=local`, `SMTP_HOST=mailpit`, port
1025). Captured email verifies the local flow, not external deliverability. For
production, disable this profile and configure your SMTP provider:

```dotenv
COMPOSE_PROFILES=
SMTP_HOST=smtp.example.com
SMTP_PORT=465
SMTP_SECURE=true
SMTP_USER=your-smtp-user
SMTP_PASSWORD=your-smtp-password
EMAIL_FROM=Your App <login@example.com>
```

Use port 587 with `SMTP_SECURE=false` for a provider requiring STARTTLS. Verify
your sending domain, then test delivery, link expiry, invitation acceptance,
logout and member removal at the deployed URL. Authentication and bridge secrets
stay server-side. Do not rotate `BETTER_AUTH_SECRET` without planning session
invalidations; change `AUTH_INTERNAL_SECRET` in both app containers together.

## Enable optional billing

Create a recurring Polar product, then set:

```dotenv
BILLING_ENABLED=true
POLAR_ENVIRONMENT=sandbox
POLAR_ACCESS_TOKEN=your-organization-access-token
POLAR_PRODUCT_ID=your-recurring-product-id
```

Use the same Polar environment in both applications. Core pages do not require
billing. Administrators can start checkout and open the customer portal.
Subscription status comes from Polar's customer-state API on request; there is
no required webhook replica. Provider downtime is reported as an error.

Before accepting real payments, use `POLAR_ENVIRONMENT=production` and verify a
real checkout, portal, cancellation and renewal. A local build or mocked test
does not verify those external flows.

## Deploy on one server

Use a Linux host with Docker and Compose. Point DNS to it and allow inbound ports
80 and 443. Configure `.env` before starting:

```dotenv
APP_BASE_URL=https://app.example.com
SITE_ADDRESS=app.example.com
BIND_ADDRESS=0.0.0.0
HTTP_PORT=80
HTTPS_PORT=443
APP_ENV=PROD
COMPOSE_PROFILES=
```

Configure SMTP as above. Run `docker compose up --build -d --wait`. Caddy manages
HTTPS and certificate renewal; keep its data volume. The same frontend image
uses runtime domain and email configuration. Never expose Next.js or Go ports
directly: the supplied deployment expects Caddy to protect internal routes.

Check `/api/health` and `/health`, `docker compose ps -a`, and
`docker compose logs --tail=100` when startup fails. RLS runtime validation rejects
a privileged Go database role. Do not replace its credentials with the migration
owner merely to make a failing check pass.

## Backups and updates

Back up PostgreSQL before an upgrade and test restoring into a separate database:

```sh
docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > starter.dump
```

The dump includes identity and account data; store it privately. After reviewing
changes, run `./setup.sh`. Existing values remain unchanged; newly required secrets
are added. Database major upgrades and existing installations need the
[upgrade guide](docs/UPGRADING.md).

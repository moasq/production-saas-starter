# B2B SaaS Starter

A small Go + Next.js starting point for a B2B application. Organization sign-in,
team members and roles, profile settings, and optional Polar subscriptions.
PostgreSQL is the only required database service.

```sh
git clone https://github.com/moasq/production-saas-starter.git
cd production-saas-starter
./setup.sh
```

Requires Docker with Compose on Linux, macOS, or WSL. Open **http://localhost:3000**.
The script creates a private `.env` and starts the containers. Without Stytch
credentials the app shows setup instructions; it never invents an authenticated
user. Add your Stytch B2B test project credentials to `.env` and run
`docker compose up -d` to enable sign-in. No frontend rebuild is needed for secrets.

What is included:

- Go API: organization authorization, members, profile, subscription status.
- Next.js: server-managed sign-in, dashboard, team/settings interface.
- PostgreSQL: organizations and accounts; embedded migrations run before API startup.
- Optional Polar checkout and customer portal, isolated from the core app.
- Docker Compose and Caddy: one host, one domain, automatic HTTPS for public domains.

The starter has no AI, document processing, OCR, RAG, vector database, object
storage, Redis, or usage-metering service. Add your product's first business
feature inside the existing API instead of adopting another framework.

[Setup and deployment](SETUP.md) · [Development](DEVELOPMENT.md) ·
[Architecture](docs/ARCHITECTURE.md) · [Upgrade an existing installation](docs/UPGRADING.md) ·
[Verification and limits](docs/VALIDATION.md)

Go remains the backend. Next.js supplies the React interface and session boundary.
Keeping these responsibilities explicit avoids a rewrite and preserves the
starter's Go focus. Billing is opt-in; live provider setup is required before
real email authentication or payment flows can work.

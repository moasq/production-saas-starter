# Upgrade an existing installation

This revision removes document upload, OCR, AI chat, RAG, embeddings, object
storage, Redis, invoice quotas and webhook-based billing replication. Treat it as
a breaking product simplification, not a patch release. Keep your current checkout
and a verified database backup until the core flows work in a separate environment.

## Fresh installation

Use a fresh Compose project and volume. The embedded baseline creates only the
organization/account schema on ordinary PostgreSQL. No vector extension or
specialized image is needed. Migrations run before API readiness.

## Existing data

Existing organization and account rows are preserved. The new migration accepts
the supported role values and preserves historical ones. Removed-feature tables,
subscriptions and quota rows are not automatically dropped; they become unused.
Existing databases with pgvector should keep a compatible pgvector image until
those legacy objects are explicitly migrated or removed after backup. Do not
attach such a data volume to the plain PostgreSQL image without a tested plan.

Original migrations 1–9 remain byte-for-byte in `go-b2b-starter/internal/db/postgres/archive/migrations`.
The migration source starts fresh installations at baseline 10 and advances clean
legacy ledgers 1–9 directly to it without executing removed-feature migrations.
The baseline creates the core schema if absent and preserves existing rows.
A database with a dirty migration version must be repaired deliberately before
startup; never force its version just to skip an error. A destructive automatic
downgrade is not provided. Restore a tested backup for rollback.

## Configuration and runtime

Use the root `.env.example` as the new configuration contract and transfer values
manually. Do not overwrite your old secrets or data. There is one `compose.yaml`;
the old production/dependency compose files are retired. The new default Compose
project/volume names may differ from yours. Explicitly migrate or configure your
existing volumes rather than accidentally starting against an empty database.

`STYTCH_ENV` is `test` or `live`, independent of build mode. `POLAR_ENVIRONMENT`
is `sandbox` or `production`, and billing requires `BILLING_ENABLED=true`.
Public Polar token variables, browser Stytch SDK tokens, storage/LLM/OCR settings,
Redis settings, and webhook secrets are no longer used. The frontend uses
server-only runtime credentials. Review the current Stytch RBAC contract and
callback configuration before enabling sign-in.

The default PostgreSQL version stays on the supported 17.x line to avoid an
implicit database-major upgrade. Go and Next.js dependencies are upgraded with
behavioral checks. Pin changes must pass both application and fresh-container
checks; major database upgrades require their own dump/restore or pg_upgrade plan.

## Billing behavior

The UI uses a single configured recurring Polar product and the customer portal.
Current state is fetched from Polar using the Stytch organization ID as the
customer external ID. Old invoice-meter metadata and local subscription tables
no longer authorize or block core pages. Remove obsolete webhook registrations in
Polar after verifying the new deployment. Add your product's own entitlement rules
if paid status should gate a business feature.

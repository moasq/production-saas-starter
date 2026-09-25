# Upgrade an existing installation

This release replaces Stytch with self-hosted Better Auth. It is an authentication
cutover, not a transparent session migration. Rehearse it on a restored copy of
your database before changing an installation used by customers.

## Back up and rehearse

1. Keep the previous application images, configuration and verified database dump.
   Restore that dump into a separate Compose project; keep the original running.
2. Preserve your PostgreSQL major version and volume mapping. Old installations
   with pgvector objects still need a compatible image until those objects are
   deliberately migrated or removed. Never attach an existing volume to a new
   major version merely by changing the image tag.
3. Transfer settings into the current `.env.example`; retain the existing database
   owner password. Generate independent application/auth database passwords,
   `BETTER_AUTH_SECRET`, and `AUTH_INTERNAL_SECRET`. Set the intended public URL
   and SMTP settings. Local rehearsal can use Mailpit.
4. Run the one-shot schema jobs against the restored database. Historical
   migrations 1–10 are unchanged. New migrations add provider-neutral identifiers
   and row-level security while retaining old provider IDs and existing data.
   Dirty migration ledgers fail closed and need deliberate repair.

The previous lean release retired document, OCR, AI/RAG, storage, Redis and
webhook billing domains. Their legacy tables are not automatically dropped.
Fresh installations use ordinary PostgreSQL and create only current schemas.

## Reconcile identities before importing

Stytch represents a member separately in every organization. Better Auth uses
one user with separate organization memberships. The importer combines normalized
email identities while retaining each organization's ID and membership boundary.
The original per-organization profile name is retained in the business
`legacy_full_name` snapshot when the shared Better Auth profile becomes current.
It does not copy sessions or trust old email verification; users must prove their
email again through a new magic link.

Local account rows are historical snapshots, not proof that a person is still
entitled to access. Before applying an import, compare the candidate accounts and
roles with a current Stytch organization/member export. Remove or deactivate stale
membership snapshots in the rehearsal database, resolve conflicting or missing
emails, and confirm at least one legitimate administrator per active workspace.
Review every elevated role. The importer must not be used to restore removed
users from an old backup.

The frontend includes `scripts/import-legacy-auth.mjs`. Run it with the auth
schema connection (`PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`) and a
separate `LEGACY_DATABASE_URL` owner connection for reading the legacy business
rows. It defaults to a dry run; `--apply` is required to write. Run it only in a
short-lived migration job or controlled development shell. Never add the owner
connection to the long-running frontend or backend environment. Check its output
and resolve rejected rows before proceeding.

Existing organizations keep their original external customer identity for Polar.
Do not recreate those customer identities or substitute a user's Better Auth ID.
Verify billing state for a known organization in rehearsal without taking new
payments. Unconfigured billing must remain disabled.

## Cut over and verify

Pause writes during the final backup/import window, repeat the reviewed migration
against the final snapshot, then start the new application. Stytch sessions and
outstanding login/invitation links are invalid after cutover. Tell users to request
new links; resend still-needed invitations. Do not run both identity authorities
concurrently against the same live membership data.

Verify two distinct tenants, invitations, users belonging to both organizations,
all three roles, logout, member removal, expired links, profile changes and the
last-administrator guard. Confirm the Go database role is neither superuser nor
BYPASSRLS and does not own protected tables. Also verify SMTP externally and test
production billing separately; local email capture and mocks do not prove those
services work.

## Roll back

Before accepting new writes, rollback can restore the previous images,
configuration and database backup together. After users have written data to the
new system, first export and reconcile those changes; restoring an old backup
blindly would lose them. Do not run an automatic destructive down migration or
re-enable old Stytch links against a partially migrated database. Keep backups
and the previous deployment until the new flows have been accepted.

# Verification — 25 September 2026

This revision implements issue #64 with self-hosted Better Auth 1.7.6, Next.js
16.3.6, Go 1.27.1 and PostgreSQL 17.11. Verification uses disposable local tenants
and a captured SMTP inbox. No production identity system or customer data was
changed. README branding, layout and the original dashboard asset are preserved.

## Application and database checks

- Full Go race tests and vet pass. PostgreSQL integration covers fresh schema,
  clean legacy ledgers 1–9, repeated migrations, dirty-ledger refusal, preserved
  profile/provider identifiers and Polar customer mapping.
- Tenant isolation is tested using a real non-owner, non-superuser, non-BYPASSRLS
  role. Unscoped reads/writes fail closed, cross-tenant reads/inserts/updates/deletes
  are denied, and a single pooled connection loses tenant context after both
  commit and rollback. Runtime checks reject elevated and inherited bypass roles.
- The real Better Auth schema and legacy importer pass fresh/repeat migration,
  dry-run with no writes, two-tenant identity mapping, inactive-row exclusion,
  duplicate/unknown-role refusal and repeat import. Import markers prevent an
  importer rerun from restoring a deliberately removed membership.
- Frontend lint, TypeScript, 15 behavior tests, production build and dependency
  audit pass. The public magic-link boundary rejects oversized names before
  persistence. Runtime secrets are not build arguments.

## Integrated behavior

The checked-in `scripts/test-auth.mjs` uses real Better Auth sessions, Go endpoints,
PostgreSQL and SMTP capture. CI runs it after a fresh Compose startup. It covers
one-use magic links, two organizations, forged tenant selectors/headers, member
ID substitution, invitation resend/acceptance/wrong recipient, all three roles,
profile updates, multiple memberships, removal/reinvitation, concurrent last-admin
protection, logout cookie replay, fixed session expiry and disabled billing.
Rate limits stay enabled; the test honors server retry intervals.

Browser verification follows the actual UI: workspace signup, captured email,
authenticated dashboard, administrator invitation, logout, recipient email proof,
invitation acceptance, second workspace creation and switching between member and
admin access. Capturing an email locally is not proof of external delivery.

The deployment uses separate schema/runtime roles, one-shot migrations, a private
auth bridge and Caddy routing. Repeat setup must preserve existing secrets and
rows. The original migration files 1–10 and dashboard image are unchanged.

## Dependency evidence and limits

Frontend production audit reports no known vulnerabilities. Pinned govulncheck
1.8.0 reports zero affected symbols and zero affected imported packages. The
required module graph still includes advisory GO-2026-5932 in the unimported
`golang.org/x/crypto/openpgp` package; no upstream fixed version is available and
the finding is not suppressed. This is not a claim that every module is free of
known vulnerabilities.

Live external SMTP delivery, Polar payment/portal lifecycle, public DNS/TLS, and
restoration of a customer's production database are not verified here. Review
[Upgrading](UPGRADING.md), reconcile source memberships, and rehearse the cutover
on a restored backup. Source review and passing negative tests reduce risk; they
do not establish perfect isolation or compliance certification.

The PR's remote checks are the authority for CI state; local checks do not imply
remote success. Engineering references are linked in [Architecture](ARCHITECTURE.md).

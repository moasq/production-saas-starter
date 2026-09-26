# Readiness and recovery

A running process, a usable database and a working customer journey are different
checks. The Go server exposes these unauthenticated, read-only probes through
Caddy. Responses use fixed status/check names and never include database errors,
connection strings, identities or credentials; they are marked `no-store`.

| Endpoint | Purpose | Failure behavior |
| --- | --- | --- |
| `/livez` (`/api/livez`) | Go can serve HTTP; no dependency calls | A dependency outage leaves it at HTTP 200 |
| `/readyz` (`/api/readyz`) | Go database access and runtime security/schema contract | HTTP 503 when PostgreSQL is unavailable, the migration version is wrong/dirty, the role is elevated, or tenant FORCE RLS is missing |
| `/health` | Compatibility alias for Go readiness | Same as `/readyz` |
| `/api/health` through Caddy | Next.js process response | Does not prove auth/database/email/billing readiness |

Directly inside the Go container, `/api/health` is also a legacy Go readiness
alias; Caddy intentionally preserves the existing frontend route at that path.
Use the unambiguous `/livez` and `/readyz` names for new monitoring configuration.
The readiness response reports `checks.database` and `checks.database_contract`.
For a failed contract, inspect `./setup.sh --doctor` and local migration logs;
repair the configuration or perform the reviewed recovery procedure below.

Ping and runtime checks share one two-second context deadline, including pool
acquisition and SQL queries. Client cancellation is propagated. Readiness reruns
the same expected migration, role and FORCE RLS checks used at startup so later
configuration drift is visible. It does not mutate data or repair a dirty ledger.
It checks enforced table RLS flags, not arbitrary policy definitions or full
schema drift; migration review, tenant behavior tests and the doctor remain
necessary.

The backend image's Docker `HEALTHCHECK` uses `/readyz`, so Compose startup waits
for the database contract before starting the frontend. Docker marks an already
running container unhealthy when its probe fails; `restart: unless-stopped`
restarts an exited process, not an unhealthy one. If adding an orchestrator,
use `/livez` for liveness restarts and `/readyz` for traffic admission. Do not
configure dependency readiness as a liveness restart probe.

Readiness is scoped to Go's local database needs. It does not claim that the
Next.js auth schema, auth bridge, SMTP delivery, DNS/TLS, or optional Polar is
working. Auth and membership checks still fail closed on each protected request;
SMTP or Polar failures remain operation errors instead of liveness restart
signals. Use `./setup.sh --doctor`, the isolated auth/browser acceptance suites,
and explicit provider verification before release. Never send payments or login
email from a recurring health probe.

## Backup and restore procedure

Keep backup data outside the source checkout and restrict access. It contains
customer data, password hashes, sessions and potentially other private values.
Use encrypted storage, retention and periodic restore rehearsals appropriate to
your deployment; this repository does not schedule backups or provide PITR.

1. Record the source commit, image digests, PostgreSQL major version, Compose
   project/volume names and private deployment configuration. Retain previous
   known-good images and the configuration needed to run them. Pause all writers
   (including custom jobs), leaving PostgreSQL running. Retain the original
   volume; never use `down --volumes` on a real installation.
2. With `umask 077`, take a custom-format `pg_dump` and a separate
   `pg_dumpall --roles-only` dump. Check both exit statuses and file sizes.
   An empty file or successful dump alone is not a verified backup. The exact
   commands and role handling are in [Upgrading](UPGRADING.md#postgresql-17-to-18).
3. Restore into a **different, fresh volume/project**, using a compatible
   PostgreSQL version and the intended rollback/release image. Provision the
   standard restricted roles from private config first, then run `pg_restore
   --clean --if-exists --exit-on-error` only against that new destination.
   Review any extra roles/extensions before restoring; never blindly apply a
   roles dump over the destination owner. The source remains untouched.
4. Check migration ledgers/checksums, row counts, original billing identifiers,
   table owners, runtime grants and auth-schema ownership. Test tenant isolation
   using the restricted runtime role, including zero rows without tenant
   context and denial of another tenant's rows. Run the matching release's
   one-shot migrations, then readiness, smoke and doctor checks.
5. Verify login, two tenants and relevant roles using approved rehearsal accounts,
   plus external SMTP and enabled billing separately. Do not run destructive
   synthetic acceptance scripts against restored customer data. Cut traffic
   over only after acceptance and record the backup/cutover time.

## Failed migration and rollback

If a one-shot migration exits nonzero, keep writers stopped. Preserve the logs,
failed database and pre-migration backup. A dirty migration ledger intentionally
blocks subsequent migration runs and Go readiness; do not delete the ledger,
force its version, edit applied migrations, or retry SQL blindly.

Before new writes, restore the verified pre-migration backup into a separate
fresh destination and start the previous compatible images/configuration. Keep
the failed source for diagnosis. If a repair is preferable, review the actual
partially applied database state and write an explicit forward migration; test
that repair on a restored copy before applying it to the affected deployment.
A migration rollback is not merely an image rollback when schema compatibility
changed. After accepting new writes, capture and reconcile them before restoring
an older backup or those writes would be lost.

## Executable release evidence

`./scripts/test-recovery.sh` builds this checkout into two uniquely named,
synthetic Compose projects with separate fresh volumes and private temporary
configuration. It ignores the checkout's `.env` and exported provider settings.
It checks:

- Fresh application smoke checks and the new health routes.
- Database outage: readiness 503, liveness 200, and recovery after restart.
- Elevated runtime role and missing FORCE RLS: readiness 503 until repaired.
- An actual conflicting migration: nonzero exit, dirty marker and refused retry.
- A separate-volume restore of real starter schemas, synthetic identities,
  memberships, original billing IDs and tenant RLS, followed by fresh startup.

The test deliberately makes its generated version-11 database claim version 10
so the real immutable version-11 migration fails on an existing column. This is
fault injection, not a repair procedure. The source remains dirty while the
restored copy becomes ready. Cleanup removes only the script's generated
containers, volumes, image tags and private temporary files.

Default ports are 14245/18445/18045 for the source and 14246/18446/18046 for the
restored copy. Override the corresponding `RECOVERY_HTTP_PORT`,
`RECOVERY_HTTPS_PORT`, `RECOVERY_MAIL_PORT`, `RECOVERY_RESTORE_HTTP_PORT`,
`RECOVERY_RESTORE_HTTPS_PORT`, and `RECOVERY_RESTORE_MAIL_PORT` if occupied.
The **Deployment recovery** workflow runs the rehearsal on pull requests and
main alongside the existing Go, database, browser and auth checks. This is
synthetic local/CI evidence; it does not verify a production backup, recovery
time objective, provider delivery or live payment.

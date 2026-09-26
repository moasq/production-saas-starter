# Verify integrations before a release

The maintained core uses self-hosted Better Auth/PostgreSQL, SMTP, and optional
Polar. Stytch, object storage, AI, OCR and RAG are removed; they have no release
checks and are not implied capabilities. No provider credentials are needed for CI.

Record each operation for an exact revision and deployment. Use four states:

| State | Meaning |
| --- | --- |
| `not_configured` | Required inputs are absent or the optional integration is disabled. |
| `configured` | Inputs are present; the external operation has not been demonstrated. |
| `verified` | The named operation succeeded in the named environment. |
| `failed` | Configuration is invalid or the attempted operation failed. |

A skipped operation is not verified. SMTP connection success does not establish
delivery. A sandbox product read does not establish checkout or payment success.
There is no aggregate production-ready flag.

## Repeatable local evidence

CI runs `scripts/test-auth.mjs` on a new Compose database with real Better Auth,
Go, PostgreSQL and a captured Mailpit inbox. It checks sign-in, invitation delivery
and acceptance, revocation, tenant isolation, RBAC and disabled billing. Browser
CI exercises the actual interface; Go billing tests and frontend Polar tests use
HTTP fixtures. These checks establish local behavior and wire-format contracts.
They do not contact external SMTP or Polar accounts.

## Optional read-only connectivity probes

Install locked frontend dependencies. Supply private runtime environment variables
through your shell or an ignored environment file; these commands never copy them
into source or reports. From `next_b2b_starter/`:

```sh
# No network access. Exit 2 means verification is still incomplete.
node scripts/check-providers.mjs

# Explicitly check SMTP connection/authentication; no email is sent.
node --env-file=.env scripts/check-providers.mjs --smtp

# Explicitly read the configured recurring product in Polar sandbox only.
node --env-file=.env scripts/check-providers.mjs --polar-sandbox
```

Use `--env-file=../.env` for the root configuration, replacing Compose-only hosts
such as `mailpit` with an address reachable from the host when appropriate. The
probe does not load Compose or assume its internal DNS is reachable on your host.
Both flags can be combined. Exit 0 means all **requested probes** passed, 1 means
a failure, and 2 means absent configuration or no probe selected. Even exit 0 leaves
the user journeys below unverified. The report contains revision, timestamp,
environment, operation, state and safe summaries; raw responses, tokens, URLs
containing tokens, email addresses and customer identifiers are omitted.

Polar checks refuse `POLAR_ENVIRONMENT=production`, follow no redirects and create
no checkouts, payments or subscriptions. SMTP checks use the application's TLS
and timeout settings. Nodemailer's [`verify`](https://nodemailer.com/smtp#verifying-the-configuration)
checks connection/authentication, not sender acceptance or recipient delivery.

## External SMTP journey

Use a separate staging deployment, two synthetic organizations and mailboxes owned
by the tester. Never point the destructive local auth suite at production accounts.

1. Configure the sending domain/provider and staging HTTPS URL. Run the SMTP probe.
2. Request sign-in through `/auth`. Confirm the actual recipient inbox receives
   the message and the link returns to the staging domain. Complete sign-in; retry
   the same link and an expired link and confirm rejection.
3. As an admin, invite a second mailbox. Confirm external delivery, resend, correct
   recipient acceptance and wrong-recipient rejection. Verify the member cannot
   invite, alter roles or read the second organization's data.
4. Log out, replay the old session and remove a member. Verify access is denied.
5. Use a disposable staging SMTP configuration to reproduce rejected delivery;
   the UI must report failure instead of claiming an invitation was delivered.
   Restore the valid configuration and repeat the journey.

Record pass/fail per operation. Keep email contents, magic links and session cookies
out of evidence. A missing inbox receipt leaves delivery `configured`, not verified.

## Polar sandbox journey

Create a separate [Polar sandbox](https://polar.sh/docs/integrate/sandbox) organization
and recurring test product. Set `BILLING_ENABLED=true`, `POLAR_ENVIRONMENT=sandbox`,
and the sandbox token/product in both applications. Use the provider's documented
test payment method; sandbox identities and products are separate from production.

1. Run the sandbox product probe, then sign into a synthetic admin workspace.
2. Complete hosted checkout. Verify Go returns the subscription for the same
   workspace and only the configured product. An unrelated workspace must fail
   checkout ownership verification. Members must be denied billing mutations.
3. Repeat the checkout return/verification. It must not mutate quotas, create a
   second subscription or change tenant ownership. A redirect alone is not payment
   evidence; wait for provider state and verify the Go response.
4. Open the portal, cancel and inspect the provider's effective date/status, then
   refresh the app. Check reactivation/renewal only if supported by the chosen
   provider product; otherwise record the limitation. Plan changes are not promised
   by this minimal integration.
5. Revoke a disposable token and retry billing. It must show an unavailable/error
   state, never silently downgrade to free or assert a paid entitlement. Restore
   staging configuration afterward. Verify core non-billing pages still work.

There is no webhook consumer or local entitlement replica in this starter, so
signature/duplicate/out-of-order webhook tests are **not applicable**, not passed.
If adding fulfillment later, implement and verify a durable idempotent event path
before relying on it. Subscription reads currently require Polar availability.

## Release evidence and cleanup

Store a private report with: exact Git SHA, deployment label, UTC timestamp, tester,
provider/environment, named operation, state, expected/observed result and a safe
evidence reference. Do not copy raw payloads, customer data or secrets. A fixture
result must explicitly say `local fixture`; a real sandbox result must say `sandbox`.
The repository's `reports/` directory is ignored for local evidence.

Remove only the synthetic users/organizations and sandbox subscriptions created
for the test after preserving the safe result. Review provider records before
retrying an ambiguous payment or invitation. Production charges, production mail
and public releases are separate actions; this procedure does not authorize them.

A release capability matrix must list external SMTP and Polar journeys separately.
Leave unrun operations unverified and link their remaining work. See
[local verification](VALIDATION.md) and [upgrade/restore guidance](UPGRADING.md).

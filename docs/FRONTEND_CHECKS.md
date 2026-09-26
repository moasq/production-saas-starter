# Frontend browser checks

The starter uses the current light UI, system font, Tailwind 3 and existing Radix
components. `app/globals.css` owns semantic color/radius tokens;
`tailwind.config.ts` maps those tokens into utilities. Prefer those tokens for new
components. Existing route palettes are preserved. CSS contains dormant `.dark`
tokens, but there is no theme switch or theme provider: dark OS preference is
checked to retain the supported light rendering, not reported as a dark theme.

Presentation belongs in route/components; queries, mutations and provider calls
stay in their existing hooks, repositories and server-only modules. Before adding
a component, reuse `components/ui`, inspect dependency/source provenance, check
its scripts and external requests, and keep credentials in server runtime config.
Do not import another repository's theme, brand rules or backend runtime.

## Run the production browser matrix

From the repository root, with Docker running and Node 24+ available:

```sh
npm ci --prefix tests/browser
npx --prefix tests/browser playwright install chromium
./scripts/test-browser.sh
```

Browser tooling has its own locked optional development package. It is not copied
into either application image and adds no runtime dependency. CI uses the same
runner with Chromium system dependencies installed. To use an installed Chrome
locally, set `BROWSER_EXECUTABLE` to its executable; record that distinction when
reporting results.

The runner creates a uniquely named Compose project with private temporary,
synthetic-only credentials. It never reads the repository `.env`, reuses an
existing database or enables billing. It starts the real production frontend,
Go API, PostgreSQL and local Mailpit, plus a second frontend with no auth config.
Ports default to 14250 (app), 18450 (HTTPS), 18050 (Mailpit), 14251 (setup screen).
Override `BROWSER_HTTP_PORT`, `BROWSER_HTTPS_PORT`, `BROWSER_MAIL_PORT` and
`BROWSER_SETUP_PORT` if needed. Its exit trap removes only its own resources.
Do not adapt this test runner to customer or restored production databases.

## Declared scope and acceptance

Each journey runs at **390 × 844** and **1440 × 1000**, with light and dark OS
preferences. Browser contexts are isolated. All account names and emails are
synthetic; magic links are delivered only to this run's Mailpit.

| Surface | State and behavior |
| --- | --- |
| `/` | Public landing, one main/h1, named CTA |
| `/auth` | Signed out, labeled email, protected dashboard redirects here |
| `/signup` | Empty and invalid fields, keyboard account/organization steps, focus transfer, pending submission, real email success |
| `/authenticate?error=invalid` | Invalid-link recovery |
| `/invite` | Missing-invitation recovery |
| `/auth` on unconfigured instance | Actual fail-closed setup screen with no sign-in inputs |
| `/dashboard` | Real signup/session/workspace, responsive navigation, mobile focus trap/Escape/restoration |
| `/dashboard/settings` | Overview, disabled billing absent, real saved profile, current team and labeled invitation dialog |

The matrix checks document headers from rendered responses, reflow, accessible
labels, one main/h1, absence of nested controls, keyboard reachability and visible
focus. It captures each declared settled state (56 captures for eight tests).
It does not claim comprehensive WCAG conformance, pixel-baseline comparisons,
Firefox/WebKit coverage, live email, invitations sent to outside recipients or
Polar payments. Go/auth integration tests own the broader authorization matrix.

Loading, empty, error, unauthorized and unconfigured behavior must be considered
for every changed surface; extend the matrix when adding a new user-visible state.
An empty workspace list and empty signup form differ from an empty team: the
workspace always retains its administrator. Never manufacture an impossible state
to make a screenshot appear complete.

## Review current evidence

`tests/browser/playwright-report/` contains the HTML report;
`tests/browser/test-results/` contains screenshots and adjacent JSON receipts.
These directories are ignored. CI retains synthetic evidence for seven days.
Each capture records Git revision, full source fingerprint, exact frontend image
ID, route/state, viewport, requested color preference, rendered theme, timestamp
browser version and screenshot hash. The source fingerprint includes uncommitted changes, and the
runner rejects changes made after the image build. A previous screenshot cannot
prove a later build. Cookies/session storage are not exported, and traces/video are disabled. Failure
messages can include one-time fixture URLs, which is another reason to use only
synthetic accounts and retain artifacts briefly.

Inspect captures at both widths before declaring visible work complete. Check
hierarchy, clipped/overlapping text, form labels, real content lengths, modal
placement, navigation and action visibility. Keep the established design; fix
observed defects without adding unrelated visual effects. Screenshots establish
rendering; the accompanying assertions establish the specific interactions.

Run frontend lint/typecheck/tests/build (`pnpm --dir next_b2b_starter verify`) and
this matrix after visible changes. A green screenshot alone does not establish
signup, invitation, checkout or portal completion.

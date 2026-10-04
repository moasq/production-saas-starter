# Frontend browser and tooling contract

Decision date: October 4, 2026. Scope: issue [#35](https://github.com/moasq/production-saas-starter/issues/35).

Keep the existing presentation and compatible CSS/compiler lines while upgrading
security patches separately. A framework or tooling major needs its own migration
and acceptance evidence. This decision does not claim that retained versions are
the latest or that all of them still receive upstream maintenance.

## Browser contract

The compatibility target follows the current
[Next.js browser baseline](https://nextjs.org/docs/app/getting-started/installation#supported-browsers):

| Browser | Minimum compatibility target |
| --- | --- |
| Chrome / Edge | 111 |
| Firefox | 111 |
| Safari | 16.4 |

New frontend work must not knowingly require a newer browser without updating
this contract and checking the affected flows. These are compatibility targets,
not a recommendation to run an obsolete browser. Use a currently maintained
browser release in production. Internet Explorer and legacy embedded webviews
are outside this contract.

Current automated evidence covers only the locked Playwright Chromium at narrow
and desktop widths, with both OS color preferences retaining the supported light
UI. Firefox, Safari/WebKit and the minimum versions above have **not** been
validated by that suite. See [frontend checks](../FRONTEND_CHECKS.md) for exact
journeys and evidence limits. Before raising a browser minimum, record a deliberate
support decision and exercise affected journeys in Chromium, Firefox and WebKit;
do not infer coverage from a successful CSS build.

## Toolchain decisions

Exact patch versions remain in the manifest and lockfile; this table defines the
accepted release lines, not a second version authority.

| Tool | Current decision | Reason and next step |
| --- | --- | --- |
| Tailwind CSS 3.4 + tailwind-merge 2.6 | Retain as a pair | Preserve the current tokens/utilities and Firefox target; migrate both together if the browser decision changes. |
| PostCSS 8 + Autoprefixer 10 + tailwindcss-animate 1 | Retain with Tailwind 3 | Current PostCSS and animation configuration belongs to this CSS stack. |
| TypeScript 5.9 | Retain the tested compiler; evaluate 6 separately | The locked typescript-eslint 8.70.1 accepts `>=4.8.4 <6.1.0`. TypeScript 7 is outside that range; 6 is eligible for a migration trial, not established as incompatible. |
| ESLint 9 | Temporary compatibility exception, **upstream EOL** | Locked/latest React 7.37.5 and jsx-a11y 6.10.2 plugins exclude ESLint 10. Resolve the lint migration before the next versioned revival release. |
| Node 24 LTS declarations | Keep `@types/node` on 24 | Node 24 is the documented development baseline; Node 26 is an additional CI/runtime target. Newer declarations must not silently admit APIs absent on Node 24. |
| React 19 and Next 16 | Keep React/DOM and their declarations aligned; match Next and eslint-config-next | Minor/patch updates use the existing reviewed group; majors need a separate migration. |

The manifest's Node 22.18 engine floor allows native TypeScript test stripping;
it does not mean every Node version above it is tested. CI validates pinned Node
24 and 26, using the pinned pnpm version. Declaration packages describe APIs;
they neither install nor polyfill a runtime. See
[DefinitelyTyped's version policy](https://github.com/DefinitelyTyped/DefinitelyTyped#how-do-definitely-typed-package-versions-relate-to-versions-of-the-corresponding-library).

### CSS migration gate

[Tailwind's upgrade guide](https://tailwindcss.com/docs/upgrade-guide) requires
Firefox 128 for v4 and recommends v3.4 when older browsers remain required.
The [tailwind-merge maintainer](https://github.com/dcastil/tailwind-merge) directs
Tailwind 3 users to the 2.6 line. Agentic Ship's Tailwind 4 instructions therefore
cannot be copied into this project's Tailwind 3 configuration.

If adopting v4, first revisit the browser target, then change the PostCSS plugin
to `@tailwindcss/postcss`, CSS entry directives/tokens, `tailwind.config.ts`,
`components.json`, animation integration and tailwind-merge in one reviewed PR.
Audit utilities whose meaning changes, including borders, rings, outlines,
shadows and radius. Preserve the visible design and existing assets. Inspect
390px/1440px rendering and keyboard/focus states, including both OS preferences,
after the complete migration; current `.dark` tokens do not promise a dark theme.
Run `pnpm verify`, dependency audits and the disposable production browser suite.

The isolated Tailwind 4 bump in [PR #93](https://github.com/moasq/production-saas-starter/pull/93)
already fails its [CI run](https://github.com/moasq/production-saas-starter/actions/runs/36871896771):
the old `darkMode` tuple fails type checking and the unchanged PostCSS plugin
fails the build. [PR #89](https://github.com/moasq/production-saas-starter/pull/89)
upgrades only tailwind-merge to 3. Neither is the coordinated migration above.

### Compiler and lint migration gate

Check the **resolved** Next lint plugins' peer ranges, not just the broad range
in eslint-config-next. The current [typescript-eslint support range](https://typescript-eslint.io/users/dependency-versions/)
allows TypeScript 6 and ESLint 10, but other plugins can impose stricter limits.
For a TypeScript 6 trial, review the [official migration notes](https://devblogs.microsoft.com/typescript/announcing-typescript-6-0/),
including the existing `baseUrl`, changed inference and declaration defaults.
Keep strict checks, API generation, Next route types and the full verify suite;
do not bypass peer checks, disable lint rules or hand-edit generated API types to
make a new compiler install. Evaluate TypeScript 7 only after the parser and
Next tooling explicitly support it.

[ESLint 9 reached EOL on August 6, 2026](https://eslint.org/version-support/).
The migration is tracked in [issue #94](https://github.com/moasq/production-saas-starter/issues/94).
Passing audits or retaining a dev-only package does not restore upstream support.
The repository maintainers must review ESLint 10 and its React/accessibility
plugins at each weekly dependency review and resolve this exception before the
next versioned release. Prefer compatible plugin releases; if they remain blocked,
evaluate an explicitly reviewed lint configuration migration that preserves the
existing rule coverage. Do not silently remove accessibility or React checks.
ESLint major Dependabot proposals stay visible to keep this work actionable.

The declaration-only [PR #87](https://github.com/moasq/production-saas-starter/pull/87)
passes its existing CI, but adopting Node 26 declarations requires a separate
runtime-baseline decision. Retain Node 24 declarations until then.

## Update policy

Dependabot groups CSS minor/patch updates and ignores **major version updates**
for Tailwind CSS, tailwind-merge, TypeScript and Node declarations. These narrow
rules make the migrations above manual; they do not freeze compatible patches.
Use the documented [Dependabot ignore/group options](https://docs.github.com/en/code-security/reference/supply-chain-security/dependabot-options-reference)
and inspect the actual update jobs; this configuration is not evidence that a bot
can produce every required fix. Keep security alerts and the scheduled full dependency audit active. If a security
fix requires a deferred major, prioritize that migration and its compatibility
checks rather than declaring the alert resolved or suppressing it.

Revisit this decision weekly and before a versioned release. A migration PR must
update this document and its corresponding Dependabot rule together, regenerate
the lockfile with the pinned pnpm, and record checks for the exact revision.

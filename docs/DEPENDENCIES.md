# Dependency maintenance

Use lockfiles and the versions checked into the repository. Do not install
`@latest` generators or regenerate files with a different local toolchain.

| Component | Version authority | Verification |
| --- | --- | --- |
| Go | `go-b2b-starter/go.mod`; match the Docker builder version | Go race tests, vet, govulncheck and container smoke |
| Go formatting | Makefile reuses the exact Dockerfile compiler image | `make -C go-b2b-starter fmt-check` |
| SQLC | `SQLC_IMAGE` in the Go Makefile | `make -C go-b2b-starter sqlc-check` checks two clean generations and the committed output |
| Node | Frontend Dockerfile and CI Node matrix | Node 24 LTS and Node 26 run lint, types, behavior tests and production builds |
| pnpm | Frontend `packageManager`; match Dockerfile and CI setup | Frozen `pnpm-lock.yaml` installation |
| Frontend libraries | Exact `package.json` entries and `pnpm-lock.yaml` | `pnpm verify`, public-secret checks and browser journeys |
| Playwright | `tests/browser/package.json` and `package-lock.json` | `npm ci --prefix tests/browser`, matching Chromium and browser CI |
| Containers | Digest-pinned Dockerfiles and `compose.yaml` | Fresh Compose migrations/auth/smoke and upgrade/restore tests |
| CI actions | Commit SHA in each workflow | Review the corresponding upstream release and workflow run |

The frontend engine floor is Node 22.18 for native TypeScript stripping in tests;
the maintained CI configurations are Node 24 and 26. Prefer the CI versions for
development. The runtime is chosen deliberately; do not infer compatibility from
an `engines` range alone. Swagger and the former hot-reload tool image were removed;
the Go-owned OpenAPI file generates the frontend contract with locked
`openapi-typescript` and `pnpm api:check` verifies it.

## Update cadence

Dependabot checks Go, the frontend, browser tooling, Dockerfiles, Compose and
GitHub Actions weekly. Next/React and their declarations move together for minor
and patch updates; Radix and Better Auth have separate small groups. Major updates,
the Polar SDK and other ungrouped dependencies remain individual review units.
SQLC and govulncheck versions inside build commands require a manual weekly review;
Dependabot does not promise to discover arbitrary tool versions in scripts.

The weekly **Scheduled dependency audit** checks reachable Go vulnerabilities and
high/critical advisories in both npm dependency trees, including developer tools.
It also runs manually and on changes to its workflow or Dependabot configuration.
Download, registry or audit errors fail the job; no step ignores errors or marks
an unavailable registry as current. Regular PR CI retains application audits and
behavior checks. A successful audit means no matching advisory was reported at
that time, not that every dependency is latest or free of defects.

Before merging an update:

1. Read the official release/migration notes. Treat auth, billing, database,
   compiler and CSS majors as compatibility work with their own issue.
2. Update linked pins together (Go module/builder, Node Dockerfile/CI matrix,
   pnpm manifest/Dockerfile/CI, or Next/React/types). Commit regenerated lockfiles.
3. Run the relevant checks above. Container/runtime changes need fresh Compose
   and migration preservation; auth/billing changes need negative behavior tests.
   HTTP fixtures do not establish successful live provider transactions.
4. Review the exact PR revision and all required checks. Never auto-merge a
   version bump merely because installation succeeded.

See GitHub's [Dependabot options reference](https://docs.github.com/en/code-security/reference/supply-chain-security/dependabot-options-reference)
for grouping behavior. Review update-job failures in the repository's dependency
graph and failed scheduled audits in Actions before claiming the maintenance run
completed. No bot-generated updates or successful future scheduled run are implied
by committing this configuration.

The explicit development pin for `yaml` selects patched 2.8.3 for the optional
Tailwind/PostCSS peer (GHSA-48c2-rrv3-qjmp). This keeps the existing YAML config
capability while replacing the vulnerable transitive 2.8.1 resolution. It is not
an application runtime dependency. Review and remove the explicit pin when its
parent dependency guarantees a patched resolution without it.

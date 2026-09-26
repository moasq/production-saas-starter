# Container platforms

The supported container targets are **Linux AMD64** (Intel/AMD servers and PCs)
and **Linux ARM64** (including Apple Silicon with Docker Desktop's Linux VM).
Windows development uses WSL2 and Linux containers as described in [Setup](../SETUP.md).
Native Windows containers, 32-bit ARM and other CPU targets are not verified.

Local development and production use the same `compose.yaml` and Dockerfiles;
there is no separate development image or architecture-specific Compose override.
Compose selects the Docker daemon's native platform. The local Mailpit profile
adds email capture; production uses external SMTP. Both application images run
as non-root users. The normal setup command stays `./setup.sh`.

The Go compiler runs on BuildKit's `BUILDPLATFORM` and receives the requested
`TARGETOS` and `TARGETARCH`. Its static binary is copied into the matching runtime
image. The Next.js build runs on the target platform so its native Node packages
match the runtime. All pinned base/Compose image digests must remain multi-platform
manifest indexes when dependencies are updated. Do not replace them with an
AMD64-only or ARM64-only image digest or hard-code a `platform:` in Compose.

## Verification

`Container portability` CI uses separate native `ubuntu-24.04` (AMD64) and
`ubuntu-24.04-arm` (ARM64) runners. Each clean runner builds both production
images from source, checks their architecture, initializes a fresh PostgreSQL
volume, runs both migration jobs, and exercises public pages, readiness and
the unauthenticated/private-bridge boundaries through Caddy. It then exports the
opposite architecture's Go binary and checks its ELF CPU header. This verifies
cross-compilation without pretending an exported binary was run.

To run the same check locally with Docker running:

```sh
./scripts/test-portability.sh
```

The script uses an isolated synthetic project, billing disabled and no external
mail. It builds sequentially, preserves existing projects and volumes, and removes
only its own containers, volumes and image tags when finished. Ports default to
14240, 18440 and 18040; set `PORTABILITY_HTTP_PORT`, `PORTABILITY_HTTPS_PORT` and
`PORTABILITY_MAIL_PORT` if needed. Docker must run on the same machine as the
script because HTTP smoke checks use localhost. An optional `amd64` or `arm64`
argument asserts the daemon's native architecture and refuses a mismatch.

The existing source/test workflows check Go behavior, vet, SQLC and frontend
lint/types/tests. This additional matrix covers container compatibility and
fresh installation, not external SMTP delivery, Polar transactions, browser
visual testing, or host-native `go run`/`pnpm dev` on every operating system.

For a deployable image of a selected target, use Docker's standard platform flag:

```sh
docker build --platform linux/arm64 -t starter-backend:arm64 go-b2b-starter
docker build --platform linux/arm64 -t starter-frontend:arm64 next_b2b_starter
```

Build on a native builder for that target when possible. A foreign-platform
Next.js build and runtime-image package installation need emulation or a native
remote builder, even though the Go compilation itself does not. Multi-platform
publishing to a registry is not part of setup or this test.

References: [Docker multi-platform builds](https://docs.docker.com/build/building/multi-platform/),
[automatic platform arguments](https://docs.docker.com/build/building/variables/#multi-platform-build-arguments),
and [GitHub's native runner labels](https://docs.github.com/en/actions/reference/runners/github-hosted-runners).

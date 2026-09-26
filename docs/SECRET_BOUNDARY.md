# Runtime secrets and build checks

Polar tokens, SMTP passwords, database passwords and auth secrets are server-only
runtime configuration. The frontend Dockerfile has no secret build arguments;
Compose supplies credentials when containers start. `.dockerignore` excludes
private environment files. Do not place credentials in `next.config`'s `env` map:
that map can inline values regardless of their names. Next.js also inlines
[`NEXT_PUBLIC_` variables](https://nextjs.org/docs/app/guides/environment-variables)
into browser bundles at build time.

From `next_b2b_starter/`, run:

```sh
node --test scripts/check-public-secrets.test.mjs
node scripts/check-public-secrets.mjs --build
```

The first check rejects secret-like public variable names in tracked application
source, Docker/Compose inputs and environment examples. The second builds with
synthetic values for runtime secrets, then scans generated browser/server output,
prerendered HTML and public assets for those values. A missing build fails rather
than silently skipping the check. Only paths and variable names appear in errors.
CI also checks the frontend image configuration and build history after building
with synthetic credentials in the host environment.

These are regressions for known configuration/bundling paths, not a general secret
scanner or proof that runtime responses cannot leak a secret. Inspect changed
server actions and responses in review. Never use real credentials as test markers
or upload private environment files or build logs containing them. Child build and
Docker output is captured rather than printed: on failure, reproduce locally with
synthetic configuration and inspect diagnostics privately.

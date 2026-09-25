#!/usr/bin/env sh
set -eu
starter_url=${STARTER_URL:-http://localhost:3000}
curl --fail --silent --show-error "$starter_url/api/health" >/dev/null
curl --fail --silent --show-error "$starter_url/health" >/dev/null
for path in / /auth /signup; do
  curl --fail --silent --show-error "$starter_url$path" >/dev/null
done
# A fresh deployment must never invent an authenticated organization.
starter_status=$(curl --silent --output /dev/null --write-out '%{http_code}' "$starter_url/api/auth/profile/me")
case "$starter_status" in
  401|403|503) ;;
  *) echo "Protected endpoint failed closed check: HTTP $starter_status" >&2; exit 1 ;;
esac
starter_internal=$(curl --silent --output /dev/null --write-out '%{http_code}' "$starter_url/internal/auth/session")
[ "$starter_internal" = "404" ] || { echo "Internal auth bridge is exposed." >&2; exit 1; }
printf '%s\n' "Frontend, API, public routes, private bridge, and unauthenticated boundary passed."

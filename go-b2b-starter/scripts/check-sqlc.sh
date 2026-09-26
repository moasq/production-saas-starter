#!/usr/bin/env sh
# Regenerate from SQL twice without changing the working tree or requiring Git.
set -eu
cd "$(dirname "$0")/.."
sqlc_image=${1:?Pass the pinned SQLC image from the Makefile}
sqlc_source="$PWD/internal/db/postgres/sqlc"
sqlc_tmp=$(mktemp -d)
trap 'rm -rf "$sqlc_tmp"' EXIT
trap 'exit 1' HUP INT TERM
for generation in first second; do
  mkdir "$sqlc_tmp/$generation"
  cp "$sqlc_source/sqlc.yml" "$sqlc_tmp/$generation/"
  cp -R "$sqlc_source/query" "$sqlc_source/migrations" "$sqlc_tmp/$generation/"
  docker run --rm --user "$(id -u):$(id -g)" \
    -v "$sqlc_tmp/$generation:/src" -w /src "$sqlc_image" generate
done
if ! diff -ru "$sqlc_source/gen" "$sqlc_tmp/first/gen"; then
  echo 'SQLC output is stale. Run make sqlc and review the generated changes.' >&2
  exit 1
fi
if ! diff -ru "$sqlc_tmp/first/gen" "$sqlc_tmp/second/gen"; then
  echo 'SQLC generation is not deterministic.' >&2
  exit 1
fi
echo 'SQLC output matches the source SQL and is deterministic across two clean generations.'

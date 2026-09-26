#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  printf 'Go files need gofmt (this check does not edit files):\n%s\n' "$unformatted" >&2
  exit 1
fi
echo 'Go formatting passed.'

#!/usr/bin/env bash
set -euo pipefail
: "${GOCACHE:?owned GOCACHE required}" "${GOMODCACHE:?owned GOMODCACHE required}" "${GOTMPDIR:?owned GOTMPDIR required}"
export GOWORK=off GOTOOLCHAIN=local GOFLAGS=''
scratch="$(mktemp -d "$GOTMPDIR/confluent-coverage.XXXXXX")"
cleanup() {
  command_status=$?
  trap - EXIT
  if ! rm -rf "$scratch"; then
    printf '%s\n' 'Failed to remove owned coverage scratch' >&2
    if [ "$command_status" -eq 0 ]; then command_status=1; fi
  fi
  exit "$command_status"
}
trap cleanup EXIT
architecture="$(go env GOARCH)"
version="$(go env GOVERSION)"
"${GOLIB:-golib}" config show --json > "$scratch/policy.json"
arguments=(-policy "$scratch/policy.json" -arch "$architecture" -version "$version")
if [ "${1:-}" = inputs ]; then
  go run scripts/check-coverage-exception.go "${arguments[@]}" -inputs
  exit
fi
# Refuse stale scope before running tests. Profile is fresh and owner-private.
go run scripts/check-coverage-exception.go "${arguments[@]}"
go test -p 2 ./... -count=1 -timeout=20m -covermode=atomic -coverpkg=github.com/faustbrian/go-schema-registry/providers/confluent/v3 -coverprofile="$scratch/fresh.out"
go run scripts/check-coverage-exception.go "${arguments[@]}" -profile "$scratch/fresh.out"

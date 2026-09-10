#!/usr/bin/env bash
# Coverage gate: fails if unit coverage (excluding test infrastructure)
# drops below the floor in scripts/coverage-floor.txt.
#
# The ignored paths mirror codecov.yml. Ratchet the floor up as the
# coverage plan (docs/coverage-plan.md, local only) progresses.
set -euo pipefail

cd "$(dirname "$0")/.."

FLOOR_FILE=scripts/coverage-floor.txt

# Reuse an existing profile if provided (e.g. coverage.txt from CI).
if [[ -n "${COVERAGE_PROFILE:-}" && -f "${COVERAGE_PROFILE:-}" ]]; then
  PROFILE="$COVERAGE_PROFILE"
else
  PROFILE=$(mktemp)
  trap 'rm -f "$PROFILE"' EXIT
  go test -coverprofile="$PROFILE" ./... >/dev/null
fi

# Sum statements from the raw profile, excluding test infra and main.go.
# (The "total:" line of `go tool cover -func` covers ALL files, so it
# cannot be used here — we must filter the profile itself.)
percent=$(awk 'NR > 1 &&
  $1 !~ /internal\/(tests|cli\/e2e|mcp\/e2e|testutil|testcompose|testcover|mcprun)\// &&
  $1 !~ /cmd\/tandoor\/main\.go/ {
  split($1, pos, ":"); split(pos[1], p, " ");
  n += $2; c += ($3 > 0 ? $2 : 0)
} END {
  printf "%.1f", 100 * c / n
}' "$PROFILE")

floor=$(cat "$FLOOR_FILE")

echo "unit coverage: ${percent}% (floor ${floor}%)"

# Compare as fixed-point ints (percentages have one decimal place).
pct10=${percent%.*}0
pct10=${pct10%?}
pct10=${pct10}${percent#*.}
floor10=${floor%.*}0
floor10=${floor10%?}
floor10=${floor10}${floor#*.}

if (( pct10 < floor10 )); then
  echo "FAIL: coverage ${percent}% is below floor ${floor}%" >&2
  exit 1
fi
echo "coverage gate passed"

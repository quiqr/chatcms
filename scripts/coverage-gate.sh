#!/usr/bin/env bash
# coverage-gate.sh <coverage.out>
#
# The gate /mip:ship enforces: >=70% statement coverage overall and >=80% on
# every core package. Core packages are the ones whose breakage is silent and
# expensive: the Quiqr model, the form round-trip, the policy function and the
# git sync engine.
set -euo pipefail

PROFILE="${1:?usage: coverage-gate.sh <coverage.out>}"
OVERALL_MIN="${CHATCMS_COVERAGE_MIN:-70}"
CORE_MIN="${CHATCMS_CORE_COVERAGE_MIN:-80}"

CORE_PACKAGES=(
  "github.com/quiqr/chatcms/internal/model"
  "github.com/quiqr/chatcms/internal/forms"
  "github.com/quiqr/chatcms/internal/policy"
  "github.com/quiqr/chatcms/internal/gitsync"
)

if [[ ! -s "$PROFILE" ]]; then
  echo "coverage-gate: $PROFILE is missing or empty" >&2
  exit 1
fi

report() {
  awk '
    /^mode:/ { next }
    {
      split($1, a, ":")
      file = a[1]
      n = split(file, p, "/")
      dir = p[1]
      for (i = 2; i < n; i++) dir = dir "/" p[i]
      stmts[dir] += $2
      total += $2
      if ($3 + 0 > 0) { covered[dir] += $2; totalcovered += $2 }
    }
    END {
      for (d in stmts) printf "%s %.1f %d\n", d, 100 * covered[d] / stmts[d], stmts[d]
      if (total > 0) printf "TOTAL %.1f %d\n", 100 * totalcovered / total, total
      else print "TOTAL 0.0 0"
    }
  ' "$PROFILE"
}

REPORT="$(report)"
echo "$REPORT" | sort

FAILED=0

OVERALL="$(echo "$REPORT" | awk '$1 == "TOTAL" { print $2 }')"
if awk -v c="$OVERALL" -v m="$OVERALL_MIN" 'BEGIN { exit !(c + 0 < m + 0) }'; then
  echo "coverage-gate: overall ${OVERALL}% is below the ${OVERALL_MIN}% minimum" >&2
  FAILED=1
fi

for pkg in "${CORE_PACKAGES[@]}"; do
  line="$(echo "$REPORT" | awk -v p="$pkg" '$1 == p { print $2 }')"
  if [[ -z "$line" ]]; then
    echo "coverage-gate: core package $pkg has no covered statements yet" >&2
    continue
  fi
  if awk -v c="$line" -v m="$CORE_MIN" 'BEGIN { exit !(c + 0 < m + 0) }'; then
    echo "coverage-gate: core package $pkg at ${line}% is below the ${CORE_MIN}% minimum" >&2
    FAILED=1
  fi
done

if [[ "$FAILED" -ne 0 ]]; then
  exit 1
fi

echo "coverage-gate: passed (overall ${OVERALL}%, core minimum ${CORE_MIN}%)"

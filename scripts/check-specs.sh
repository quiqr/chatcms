#!/usr/bin/env bash
# check-specs.sh <change-name>
#
# A change's proposal declares which capabilities it introduces or modifies.
# `nix flake check` cannot notice a spec file that was never written, and
# `openspec validate` is satisfied by a single delta, so a change can ship with
# a capability it only promised. This closes that gap: every capability path in
# the proposal must have a spec file in the change.
set -euo pipefail

CHANGE="${1:?usage: check-specs.sh <change-name>}"

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

DIR="openspec/changes/${CHANGE}"
PROPOSAL="${DIR}/proposal.md"

if [[ ! -f "$PROPOSAL" ]]; then
  echo "check-specs: no proposal at $PROPOSAL" >&2
  exit 1
fi

# Capability paths are the backticked paths in the bullets under the
# Capabilities heading, up to the next top-level heading.
CAPABILITIES="$(
  awk '
    /^## Capabilities/ { inside = 1; next }
    /^## / { inside = 0 }
    inside && /^- `/ {
      line = $0
      sub(/^- `/, "", line)
      sub(/`.*$/, "", line)
      print line
    }
  ' "$PROPOSAL"
)"

if [[ -z "$CAPABILITIES" ]]; then
  echo "check-specs: ${CHANGE} declares no capabilities"
  exit 0
fi

MISSING=0
while IFS= read -r capability; do
  [[ -z "$capability" ]] && continue
  if [[ ! -f "${DIR}/specs/${capability}/spec.md" ]]; then
    echo "check-specs: ${CHANGE} declares the capability ${capability} but has no ${DIR}/specs/${capability}/spec.md" >&2
    MISSING=1
  fi
done <<< "$CAPABILITIES"

if [[ "$MISSING" -ne 0 ]]; then
  echo "check-specs: write the missing spec, or drop the capability from the proposal" >&2
  exit 1
fi

echo "check-specs: every capability ${CHANGE} declares has a spec"

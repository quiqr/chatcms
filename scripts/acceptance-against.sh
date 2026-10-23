#!/usr/bin/env bash
# acceptance-against.sh <repository> [spec ...]
#
# Runs the acceptance suite against a real site instead of the fixture one.
# The gate runs it against the fixture, because a build sandbox may not reach
# the network and a check that depends on somebody else's server is a check
# that goes red on their bad mornings. This is how the same specs are run
# against the customer's own repository before a release.
#
#   bash scripts/acceptance-against.sh git@github.com:korenaar/site.git
#   bash scripts/acceptance-against.sh /path/to/a/clone 03-forms-and-saving
#
# A spec that names a page the real site has not got says so plainly: it is
# the fixture's page that is missing, not the product that is broken.
set -euo pipefail

REMOTE="${1:?usage: acceptance-against.sh <repository> [spec ...]}"
shift || true

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

echo "==> building the server and the seed"
nix build .#chatcmsd --no-link --print-out-paths > /tmp/chatcms-acceptance-bin
BIN="$(cat /tmp/chatcms-acceptance-bin)/bin/chatcmsd"
go build -o "$ROOT/.acceptance-seed" ./cmd/seed

export CHATCMS_BIN="$BIN"
export CHATCMS_SEED="$ROOT/.acceptance-seed"
export CHATCMS_SEED_REMOTE="$REMOTE"

echo "==> running the acceptance tests against $REMOTE"
cd test/e2e
npm install --no-audit --no-fund >/dev/null
exec npx playwright test "$@"

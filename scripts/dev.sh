#!/usr/bin/env bash
# Starts everything needed to click through ChatCMS on this machine: a local
# PostgreSQL, a master key, a seeded database with one site that has a remote
# of its own on disk, and the server. Run it inside `nix develop`.
#
#   bash scripts/dev.sh          start the server, seeding only the first time
#   bash scripts/dev.sh fresh    throw the database and the state away first
#
# Then run the frontend in another terminal:
#
#   cd web && npm run dev
#
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

state="$root/.chatcms-state"
port="${CHATCMS_PORT:-8080}"

if ! command -v postgres >/dev/null; then
    echo "dev: run this inside 'nix develop', which pins PostgreSQL, git and Hugo." >&2
    exit 1
fi

if [ "${1:-}" = "fresh" ]; then
    echo "==> throwing away the database and the state"
    bash "$root/scripts/dev-db.sh" reset >/dev/null
    rm -rf "$state"
else
    bash "$root/scripts/dev-db.sh" up >/dev/null
fi

url="postgres:///chatcms_dev?host=$root/.pgdata/sock"

mkdir -p "$state"
key="$state/master.key"
if [ ! -f "$key" ]; then
    echo "==> making a master key for development"
    head -c 32 /dev/urandom > "$key"
    chmod 600 "$key"
fi

export CHATCMS_DATABASE_URL="$url"
export CHATCMS_SECRET_KEY_FILE="$key"
export CHATCMS_STATE_DIR="$state"
export CHATCMS_LISTEN="127.0.0.1:$port"
export CHATCMS_BASE_URL="http://127.0.0.1:$port"

if [ ! -d "$state/sites" ]; then
    echo "==> seeding"
    go run ./cmd/seed "$url"
fi

cat <<NOTICE

==> ChatCMS on http://127.0.0.1:$port
    sign in as owner@example.org with: a long enough passphrase

    korenaar      has a site, unpublished work, and somewhere to publish to
    studio-noord  has no site, so the setup screen has something to do

    The frontend in another terminal, with live reloading:
      cd web && npm run dev
    It serves on http://localhost:5173 and sends /api here by itself.

NOTICE

exec go run ./cmd/chatcmsd

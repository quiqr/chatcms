#!/usr/bin/env bash
# Starts everything needed to click through ChatCMS on this machine: a local
# PostgreSQL, a master key, a seeded database with one site that has a remote
# of its own on disk, and the server. Run it inside `nix develop`.
#
#   bash scripts/dev.sh          start the server, seeding only the first time
#   bash scripts/dev.sh fresh    throw the database and the state away first
#   bash scripts/dev.sh stop     stop a server left running, and the database
#
# Ctrl-C stops the server. The database keeps running, because the next start
# is quicker with it up; "stop" puts it away as well.
#
# Then run the frontend in another terminal:
#
#   cd web && npm run dev
#
# To reach it from another machine, such as a phone on the same network, give
# it an address to listen on and the address that machine will use:
#
#   CHATCMS_HOST=0.0.0.0 bash scripts/dev.sh
#   cd web && npm run dev -- --host
#
# Anything on the network can then reach it, with a password written in this
# script, so do that on a network you trust and not on a shared one.
#
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

state="$root/.chatcms-state"
port="${CHATCMS_PORT:-8080}"
host="${CHATCMS_HOST:-127.0.0.1}"

# The address a browser will use. It is what invitation links are written
# with, so on a network it has to be the address that machine can reach
# rather than this one's idea of itself.
reachable="${CHATCMS_REACHABLE_HOST:-$host}"
if [ "$reachable" = "0.0.0.0" ] || [ "$reachable" = "::" ]; then
    reachable="$(hostname -I 2>/dev/null | awk '{print $1}')"
    reachable="${reachable:-127.0.0.1}"
fi

if ! command -v postgres >/dev/null; then
    echo "dev: run this inside 'nix develop', which pins PostgreSQL, git and Hugo." >&2
    exit 1
fi

if [ "${1:-}" = "stop" ]; then
    if pkill -f "$state/chatcmsd"; then
        echo "==> stopped the server"
    else
        echo "==> no server of this script's was running"
    fi
    bash "$root/scripts/dev-db.sh" down >/dev/null 2>&1 && echo "==> stopped the database"

    # Something else may hold the port: a server started another way, or
    # another program altogether. Naming it is more use than killing it.
    sleep 1
    holder="$(ss -ltnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | head -1 || true)"
    if [ -n "$holder" ]; then
        echo "==> something still holds port $port: pid $holder ($(ps -p "$holder" -o comm= 2>/dev/null))"
        echo "    stop it with: kill $holder"
    fi
    exit 0
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
export CHATCMS_LISTEN="$host:$port"
export CHATCMS_BASE_URL="http://$reachable:$port"

frontendFlag=""
if [ "$host" != "127.0.0.1" ] && [ "$host" != "localhost" ]; then
    frontendFlag=" -- --host"
fi

if [ ! -d "$state/sites" ]; then
    echo "==> seeding"
    go run ./cmd/seed "$url"
fi

cat <<NOTICE

==> ChatCMS on http://$reachable:$port
    sign in as owner@example.org with: a long enough passphrase

    korenaar      has a site, unpublished work, and somewhere to publish to
    studio-noord  has no site, so the setup screen has something to do

    The frontend in another terminal, with live reloading:
      cd web && npm run dev$frontendFlag
    It serves on http://$reachable:5173 and sends /api here by itself.

    Ctrl-C stops this. The database keeps running; to put that away too:
      bash scripts/dev.sh stop

NOTICE

# Built rather than run with "go run", which starts the server as a child of
# itself: stopping go run then leaves the server holding the port. One
# process is one thing to stop, and Ctrl-C reaches it.
echo "==> building"
go build -o "$state/chatcmsd" ./cmd/chatcmsd

exec "$state/chatcmsd"

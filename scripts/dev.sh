#!/usr/bin/env bash
# Starts everything needed to click through ChatCMS on this machine: a local
# PostgreSQL, a master key, a seeded database with one site that has a remote
# of its own on disk, and the server. Run it inside `nix develop`.
#
#   bash scripts/dev.sh          start the server, seeding only the first time
#   bash scripts/dev.sh fresh    throw this instance's database and state away
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
#   cd web && CHATCMS_HOST=0.0.0.0 npm run dev
#
# Both halves read the same setting, so a phone that can reach one can reach
# the other.
#
# Anything on the network can then reach it, with a password written in this
# script, so do that on a network you trust and not on a shared one.
#
# A second instance, so that two people can work without treading on each
# other, needs a name. It gets its own port, its own state and its own
# database, sharing only the PostgreSQL that holds them:
#
#   CHATCMS_INSTANCE=test bash scripts/dev.sh
#   cd web && CHATCMS_PORT=8081 npm run dev -- --port 5174
#
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

# An instance is a name, or nothing for the usual one. It decides the port,
# where the state lives and which database is used, so that two of them never
# write the same working copy.
instance="${CHATCMS_INSTANCE:-}"
if [ -n "$instance" ]; then
    case "$instance" in
        *[!a-z0-9]*)
            echo "dev: an instance is named with small letters and digits, so that it can name a database too." >&2
            exit 1
            ;;
    esac
    state="$root/.chatcms-state-$instance"
    database="chatcms_dev_$instance"
    port="${CHATCMS_PORT:-8081}"
else
    state="$root/.chatcms-state"
    database="chatcms_dev"
    port="${CHATCMS_PORT:-8080}"
fi

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

export CHATCMS_DEV_DB="$database"

if [ "${1:-}" = "stop" ]; then
    if pkill -f "$state/chatcmsd"; then
        echo "==> stopped the server"
    else
        echo "==> no server of this instance was running"
    fi

    # The PostgreSQL is shared, so it is only put away when nothing else is
    # using it. Another instance's server still running is reason to leave it.
    if pgrep -f "\.chatcms-state[^/]*/chatcmsd" >/dev/null; then
        echo "==> leaving the database up; another instance is still running"
    else
        bash "$root/scripts/dev-db.sh" down >/dev/null 2>&1 && echo "==> stopped the database"
    fi

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

bash "$root/scripts/dev-db.sh" up >/dev/null

# Throwing away means this instance's database and state, never the whole
# PostgreSQL, which another instance may be using. Use dev-db.sh reset to
# throw that away on purpose.
if [ "${1:-}" = "fresh" ]; then
    echo "==> throwing away $database and $(basename "$state")"
    PGHOST="$root/.pgdata/sock" dropdb --if-exists "$database"
    PGHOST="$root/.pgdata/sock" createdb "$database"
    rm -rf "$state"
fi

url="postgres:///$database?host=$root/.pgdata/sock"

mkdir -p "$state"
key="$state/master.key"
if [ ! -f "$key" ]; then
    echo "==> making a master key for development"
    # Hex rather than raw bytes: the key file is read as text with whitespace
    # trimmed, and raw bytes can begin or end with a space.
    od -An -tx1 -N32 /dev/urandom | tr -d " \n" > "$key"
    chmod 600 "$key"
fi

export CHATCMS_DATABASE_URL="$url"
export CHATCMS_SECRET_KEY_FILE="$key"
export CHATCMS_STATE_DIR="$state"
export CHATCMS_LISTEN="$host:$port"
export CHATCMS_BASE_URL="http://$reachable:$port"

# The frontend reads the same two settings, so what to type is the same
# words in the same order rather than a second set of flags to remember.
frontendSettings=""
if [ "$host" != "127.0.0.1" ] && [ "$host" != "localhost" ]; then
    frontendSettings="CHATCMS_HOST=$host "
fi

frontendFlag=""
if [ -n "$instance" ]; then
    frontendSettings="${frontendSettings}CHATCMS_PORT=$port "
    frontendFlag=" -- --port 5174"
fi

if [ ! -d "$state/sites" ]; then
    echo "==> seeding"
    go run ./cmd/seed "$url"
fi

cat <<NOTICE

==> ChatCMS on http://$reachable:$port${instance:+  (instance: $instance)}
    sign in as owner@example.org with: a long enough passphrase

    korenaar      has a site, unpublished work, and somewhere to publish to
    studio-noord  has no site, so the setup screen has something to do

    The frontend in another terminal, with live reloading:
      cd web && ${frontendSettings}npm run dev$frontendFlag

    Ctrl-C stops this. The database keeps running; to put that away too:
      ${instance:+CHATCMS_INSTANCE=$instance }bash scripts/dev.sh stop

NOTICE

# Built rather than run with "go run", which starts the server as a child of
# itself: stopping go run then leaves the server holding the port. One
# process is one thing to stop, and Ctrl-C reaches it.
echo "==> building"
go build -o "$state/chatcmsd" ./cmd/chatcmsd

exec "$state/chatcmsd"

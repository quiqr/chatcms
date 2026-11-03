#!/usr/bin/env bash
# Brings up the old Quiqr, the application whose behaviour ChatCMS promises to
# reproduce, with a site mounted in it, so that a question about what a field
# does can be answered by the application rather than by reading its source.
#
#   bash scripts/reference.sh up        start it, setting it up the first time
#   bash scripts/reference.sh stop      stop it
#   bash scripts/reference.sh fresh     throw everything away and set up again
#   bash scripts/reference.sh record <case> <single> <field> <value>
#                                       record what one edit did, as files
#   bash scripts/reference.sh fields    say what the mounted model declares
#
# It needs the network the first time, so it is never part of `nix flake
# check`. What it records is committed, and the tests that read a recording
# run offline like every other test.
#
# Everything it writes lives under .reference-state, which is ignored.
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"
state="$root/.reference-state"

# shellcheck source=reference-pins.env
source "$root/scripts/reference-pins.env"

app="$state/app"
conf="$state/conf"
library="$state/library"
home="$state/home"
# The old application keeps a site as <library>/sites/<key>/<workspace>, and
# a workspace is any subdirectory of it that looks like a site. Cloning the
# template straight into one is what makes it mounted rather than merely
# present.
site="$library/sites/$REFERENCE_SITE_KEY/main"
log="$state/server.log"
pidfile="$state/server.pid"
recordings="$root/test/fixtures/reference"

fail() {
    echo "reference: $*" >&2
    exit 1
}

# inside runs a command in the old application's own dev shell, which is what
# the bean asks for and what supplies embgit: the application talks to git
# through it, so building our own would be a second place that knows how it
# reaches a repository. The one thing that shell does not do is put libstdc++
# on the loader path, which its prebuilt native modules need; that single
# export is the whole difference between quiqr/quiqr-desktop and
# mipmip/quiqr-desktop, and carrying it here is why upstream can be preferred.
inside() {
    nix develop "$app" --no-warn-dirty --command bash -c '
        export LD_LIBRARY_PATH="$(nix eval --raw nixpkgs#stdenv.cc.cc.lib.outPath)/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
        cd "$1"
        shift
        "$@"
    ' -- "$app" "$@"
}

says_what_it_is() {
    echo "==> the reference"
    printf '    %-12s %s\n' \
        "quiqr" "$REFERENCE_APP_COMMIT" \
        "site" "$REFERENCE_SITE_COMMIT" \
        "address" "http://127.0.0.1:$REFERENCE_PORT" \
        "user" "$REFERENCE_USER" \
        "password" "$REFERENCE_PASSWORD"
    if [ -d "$app" ]; then
        inside bash -c 'printf "    %-12s %s\n" \
            "node" "$(node --version)" \
            "hugo" "$(${HUGO_PATH:-hugo} version | cut -d" " -f2)" \
            "embgit" "${EMBGIT_PATH:-not set}"'
    fi
}

running() {
    [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null
}

stop() {
    if running; then
        kill "$(cat "$pidfile")" 2>/dev/null || true
        for _ in $(seq 1 50); do
            running || break
            sleep 0.1
        done
        running && kill -9 "$(cat "$pidfile")" 2>/dev/null || true
    fi
    rm -f "$pidfile"
}

fetch() {
    local remote=$1 commit=$2 into=$3
    if [ -d "$into/.git" ] && [ "$(git -C "$into" rev-parse HEAD)" = "$commit" ]; then
        return
    fi
    echo "==> fetching $remote at ${commit:0:8}"
    rm -rf "$into"
    mkdir -p "$into"
    git -C "$into" init --quiet
    git -C "$into" remote add origin "$remote"
    git -C "$into" fetch --quiet --depth 1 origin "$commit" \
        || fail "could not fetch $commit from $remote. The first run needs the
network, and Nix needs whatever credentials this machine uses for GitHub:
the same ones the project's own flake needs."
    git -C "$into" checkout --quiet FETCH_HEAD
}

build() {
    if [ -f "$app/packages/frontend/build/index.html" ] \
        && [ -f "$app/packages/adapters/standalone/dist/main.js" ]; then
        return
    fi
    echo "==> installing and building the old Quiqr; this takes a while, once"
    inside npm install --no-audit --no-fund
    for package in @quiqr/types @quiqr/backend @quiqr/frontend @quiqr/adapter-standalone; do
        inside npm run build -w "$package"
    done
}

# mount_the_site points the old application at a library of our own and puts
# one site in it. Without the first half it reads ~/Quiqr and edits whatever
# sites somebody already has on this machine, which is not a reference and is
# not ours to touch.
mount_the_site() {
    mkdir -p "$conf" "$library/sites/$REFERENCE_SITE_KEY" "$home"
    cat > "$conf/instance_settings.json" <<JSON
{
  "storage": {
    "type": "fs",
    "dataFolder": "$library"
  }
}
JSON
    cat > "$library/sites/$REFERENCE_SITE_KEY/config.json" <<JSON
{
  "key": "$REFERENCE_SITE_KEY",
  "name": "$REFERENCE_SITE_NAME",
  "source": { "type": "folder", "path": "main" }
}
JSON
}

make_the_user() {
    local users="$conf/users.json"
    if [ -f "$users" ] && grep -q "\"$REFERENCE_USER\"" "$users"; then
        return
    fi
    echo "==> making $REFERENCE_USER"
    printf '%s\n' "$REFERENCE_PASSWORD" \
        | QUIQR_CONF_DIR="$conf" HOME="$home" inside node packages/adapters/standalone/dist/cli/user-admin.js \
            add "$REFERENCE_USER" >/dev/null

    [ -f "$users" ] || fail "the user tool wrote no $users; the old application
may have changed where it keeps users since ${REFERENCE_APP_COMMIT:0:8}."

    # Created users must change their password at first sign-in, which would
    # make this one command and then a person. Cleared here so that setting
    # the reference up stays one command.
    python3 - "$users" <<'PY'
import json, sys
path = sys.argv[1]
with open(path) as file:
    users = json.load(file)
rows = users["users"] if isinstance(users, dict) and "users" in users else users
if not isinstance(rows, list):
    raise SystemExit(f"{path} is not the shape this script knows: {type(rows)}")
for row in rows:
    row["mustChangePassword"] = False
with open(path, "w") as file:
    json.dump(users, file, indent=2)
PY
}

start() {
    running && return
    echo "==> starting on port $REFERENCE_PORT"
    (
        cd "$app"
        QUIQR_CONF_DIR="$conf" \
        HOME="$home" \
        PORT="$REFERENCE_PORT" \
        HOST=127.0.0.1 \
        HUGO_PREVIEW_PORT="$REFERENCE_PREVIEW_PORT" \
        NODE_ENV=production \
        nohup "$root/scripts/reference-run.sh" "$app" > "$log" 2>&1 &
        echo $! > "$pidfile"
    )

    for _ in $(seq 1 120); do
        if curl -fsS -o /dev/null "http://127.0.0.1:$REFERENCE_PORT/" 2>/dev/null; then
            return
        fi
        running || fail "the server stopped while starting. Its output:
$(tail -20 "$log")"
        sleep 0.5
    done
    fail "the server did not answer on port $REFERENCE_PORT within a minute.
Its output:
$(tail -20 "$log")"
}

up() {
    mkdir -p "$state"
    fetch "$REFERENCE_APP_REMOTE" "$REFERENCE_APP_COMMIT" "$app"
    mount_the_site
    fetch "$REFERENCE_SITE_REMOTE" "$REFERENCE_SITE_COMMIT" "$site"
    build
    make_the_user
    start
    says_what_it_is
    echo
    echo "    The site is mounted from $site, which is a checkout, so"
    echo "    git -C .reference-state/site diff says what an edit did."
}

fields() {
    [ -d "$site" ] || fail "no site yet; run: bash scripts/reference.sh up"
    echo "==> what the mounted model declares, read with ChatCMS's own parser"
    go run ./cmd/reference fields "$site"
    echo "    easymde is the one type the old application draws that this"
    echo "    site does not declare. file and image appear in its"
    echo "    documentation with no component behind them."
}

record() {
    local case=$1 single=$2 field=$3 value=$4
    [ -d "$site" ] || fail "no site yet; run: bash scripts/reference.sh up"
    running || fail "the reference is not running; run: bash scripts/reference.sh up"

    local described file title
    described=$(go run ./cmd/reference single "$site" "$single") \
        || fail "the model does not describe a single called $single"
    file=$(printf '%s' "$described" | python3 -c 'import json,sys; print(json.load(sys.stdin)["file"])')
    title=$(printf '%s' "$described" | python3 -c 'import json,sys; print(json.load(sys.stdin)["title"])')

    git -C "$site" diff --quiet -- "$file" \
        || fail "$file has changes already, so a recording would hold more than
one edit. Put it back first:
  git -C $site checkout -- $file"

    # The recording is a site of its own: the model as it was and the file as
    # it was, at the paths the model names. A test can then read it the way
    # the server reads any site, with nothing of ours in between.
    local into="$recordings/$case"
    rm -rf "$into"
    mkdir -p "$into/site/$(dirname "$file")" "$into/after/$(dirname "$file")"
    cp -r "$site/quiqr" "$into/site/quiqr"
    rm -rf "$into/site/quiqr/model/select_images" "$into/site/quiqr/etalage"
    cp "$site/$file" "$into/site/$file"

    echo "==> recording \"$case\": $single.$field in the old Quiqr"
    node "$root/scripts/reference-record.mjs" \
        "$REFERENCE_PORT" "$REFERENCE_SITE_NAME" "$title" "$field" "$value" \
        || fail "the edit was not made, so there is nothing to record"

    git -C "$site" diff --quiet -- "$file" \
        && fail "$file did not change, though the form said it saved"

    cp "$site/$file" "$into/after/$file"
    cat > "$into/case.json" <<JSON
{
  "single": "$single",
  "field": "$field",
  "value": "$value",
  "file": "$file",
  "quiqr": "$REFERENCE_APP_COMMIT",
  "site": "$REFERENCE_SITE_COMMIT",
  "divergences": []
}
JSON

    echo "==> what the old Quiqr did"
    git -C "$site" --no-pager diff -- "$file"
    git -C "$site" checkout -- "$file"
    echo "==> recorded into $into"
    echo "    Anything above beyond the edit itself is a difference between"
    echo "    the two applications. Each one has to be named in case.json"
    echo "    with a reason, or the test will refuse the recording."
}

case "${1:-up}" in
    up) up ;;
    stop) stop; echo "==> stopped" ;;
    fresh)
        stop
        echo "==> throwing away $state"
        rm -rf "$state"
        up
        ;;
    fields) fields ;;
    record)
        shift
        [ $# -eq 4 ] || fail "usage: reference.sh record <case> <single> <field> <value>"
        record "$@"
        ;;
    *) fail "usage: reference.sh up|stop|fresh|fields|record" ;;
esac

#!/usr/bin/env bash
# dev-db.sh up|down|reset|psql
#
# Throwaway PostgreSQL for development and tests, living in ./.pgdata.
# Listens on a unix socket in ./.pgdata/sock so no port is claimed.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
PGDATA="$ROOT/.pgdata/data"
PGSOCK="$ROOT/.pgdata/sock"
DBNAME="${CHATCMS_DEV_DB:-chatcms_dev}"

export PGDATA PGHOST="$PGSOCK"

start() {
  mkdir -p "$PGSOCK"
  if [[ ! -d "$PGDATA" ]]; then
    initdb --username="$USER" --auth=trust --encoding=UTF8 --no-locale >/dev/null
  fi
  if ! pg_ctl status >/dev/null 2>&1; then
    pg_ctl -o "-k $PGSOCK -h ''" -l "$ROOT/.pgdata/postgres.log" start
  fi
  createdb "$DBNAME" 2>/dev/null || true
  echo "CHATCMS_DATABASE_URL=postgres:///$DBNAME?host=$PGSOCK"
}

case "${1:-up}" in
  up) start ;;
  down) pg_ctl stop || true ;;
  reset)
    pg_ctl stop 2>/dev/null || true
    rm -rf "$ROOT/.pgdata"
    start
    ;;
  psql) psql "$DBNAME" ;;
  *)
    echo "usage: dev-db.sh up|down|reset|psql" >&2
    exit 1
    ;;
esac

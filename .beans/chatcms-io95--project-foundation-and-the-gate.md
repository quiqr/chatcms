---
# chatcms-io95
title: Project foundation and the gate
status: todo
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:21Z
parent: chatcms-8vem
---

## Purpose

Turn the scaffold into a running application skeleton: a Go HTTP server with
configuration, an embedded migration runner, sqlc-generated queries, and a Vite
frontend that the gate builds. Everything after this epic assumes it exists.

## Scope

- `internal/app`: HTTP server lifecycle, graceful shutdown, structured logging.
- `internal/httpapi`: router, JSON encoding helpers, request id, panic recovery,
  `/api/health` reporting database reachability and the migration version.
- `db/migrations` with goose, embedded with `go:embed`, run on start-up before
  the listener opens. Failure to migrate is a fatal start-up error.
- `db/queries` plus `sqlc.yaml`, generating into `internal/store`.
- `web/`: Vite + React + TypeScript + Tailwind v4, built by Nix and embedded
  into the binary for production, proxied in development.
- `flake.nix`: add the `frontend` check and a `web` package with a pinned
  `npmDepsHash`; add the integration test check that runs migrations from an
  empty database.
- `scripts/dev-db.sh` proven to work end to end in the dev shell.

## Done when

- `nix flake check` builds the Go binary and the frontend, and runs both test
  suites.
- Starting the binary against an empty database creates the schema and serves
  `/api/health` with a 200.
- A second start against the migrated database is a no-op.
- The frontend dev server proxies `/api` to the Go server.

## References

- `docs/ARCHITECTURE.md`, the package table and the request path.
- `docs/BRIEFING-M01.md` section 3 for the locked stack.

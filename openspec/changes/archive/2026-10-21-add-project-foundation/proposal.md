## Why

The repository has a scaffold and a green gate but no running application: no
HTTP server, no database schema, no frontend. Every milestone 01 epic assumes
all three exist, so this is the one change that unblocks the rest. Epic
`chatcms-io95`.

It serves no single acceptance test from `docs/BRIEFING-M01.md` section 6
directly. It serves the closing paragraph of that section: the app builds and
deploys through the Nix flake, and migrations run cleanly from an empty
database.

## What Changes

- The binary becomes a server. `chatcmsd` starts an HTTP listener, serves a
  JSON API under `/api`, shuts down gracefully, and logs in a structured
  format.
- Migrations are embedded and run before the listener opens. goose migrations
  live in `db/migrations`, are embedded with `go:embed`, and a failure to
  migrate is a fatal start-up error rather than a degraded service.
- Database access is generated, not hand-written. sqlc reads `db/queries` and
  generates into `internal/store`. The generated code is committed so the gate
  does not need sqlc at build time.
- A frontend exists and is built by Nix. `web/` is a Vite + React + TypeScript
  + Tailwind v4 application. Nix builds it with a pinned `npmDepsHash`; the Go
  binary embeds the built assets and serves them in production, while
  development proxies `/api` to the Go server.
- The gate grows. `nix flake check` gains a `frontend` check and an
  `integration` check that runs the migrations against a real PostgreSQL from
  an empty database.
- `/api/health` reports whether the database is reachable and which migration
  version is applied.

## Capabilities

### New Capabilities

- `platform/server-runtime`: how the server starts, what it listens on, how it
  is configured, how it reports health, and how it stops.
- `platform/database-schema`: how the schema is defined, migrated and queried,
  including the start-up migration contract.
- `platform/frontend-delivery`: how the frontend is built, embedded and served,
  and how development differs from production.

### Modified Capabilities

None. This is the first change; there are no existing specs.

## Impact

- **New code**: `internal/httpapi`, `internal/store`, `db/migrations`,
  `db/queries`, `web/`.
- **Changed code**: `internal/app` grows from a configuration reader into the
  process lifecycle owner. `cmd/chatcmsd/main.go` is unchanged.
- **Dependencies added**: a PostgreSQL driver (pgx), goose as a library, and
  the frontend dependency tree. Each is pinned, and `vendorHash` in
  `flake.nix` stops being `null`.
- **Build**: `nix flake check` gets slower on its first run because the Node
  dependency tree is built once.
- **Seams**: none of the four seams named in `docs/ARCHITECTURE.md` are
  implemented here. `internal/httpapi` is written so that `internal/policy`
  can be inserted as the single authorisation decision point by the tenancy
  epic, which means handlers take an actor from the request context rather
  than reading a session themselves.

## Assumptions

- **No email delivery yet.** Nothing in this change sends mail, so no SMTP
  configuration is introduced.
- **No authentication yet.** Every endpoint added here is unauthenticated.
  `/api/health` stays unauthenticated afterwards; anything else added here is
  removed or protected by the authentication epic.
- **PostgreSQL 16** is the target, matching `postgresql_16` in the flake.
- **The frontend is a placeholder.** It renders enough to prove the build,
  the embed and the proxy work. The design foundation epic replaces it.
- **Network during the build.** Nix needs one network-enabled fixed-output
  derivation to fetch the Node dependency tree. If the sandbox cannot reach
  the network, the `frontend` check is the only part that cannot run, and the
  change is not done until it does.

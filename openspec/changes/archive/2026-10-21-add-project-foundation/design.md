## Context

The repository currently holds a scaffold: a Nix flake whose gate runs `go
vet`, race tests and a coverage gate, a Go binary that prints its
configuration, and documentation. There is no server, no schema and no
frontend. See `proposal.md` for why this change comes first, and
`docs/ARCHITECTURE.md` for the package layout this has to land in.

Two constraints shape everything below. Deployment is Nix-only, so whatever
the gate cannot build is not deployable. And the gate is the single command
that decides whether a change ships, so every artifact this change introduces
has to end up behind `nix flake check`, including the ones that are awkward
to run in a sandbox.

## Goals / Non-Goals

**Goals:**

- A server process with a lifecycle an operator can reason about: it either
  starts correctly or it does not start at all.
- A schema that evolves through embedded migrations, safe to run from two
  instances at once.
- A frontend that the gate builds and the binary carries, so a deployment is
  one binary.
- Room for `internal/policy` to become the single authorisation decision point
  without rewriting handlers.

**Non-Goals:**

- Any authentication, session or user concept. The authentication epic owns
  that, and this change deliberately adds no user table.
- Any real screen. The frontend here proves the build and the embed; the
  design foundation epic replaces its contents.
- Connection pool tuning, read replicas, or anything else that is a
  production concern before there is production traffic.

## Decisions

### The router is the standard library

`net/http.ServeMux` with Go 1.22 method and wildcard patterns, not chi or
gorilla. The API surface for milestone 1 is small and conventional, and the
standard mux now does path parameters and method matching. Middleware is three
small functions, written here, composed explicitly.

Alternative considered: chi. Rejected because it buys little over the current
standard library and adds a dependency to the one package every future epic
touches.

### Database access is pgx plus sqlc, with no ORM

`pgx/v5` with `pgxpool`, and sqlc generating into `internal/store` using the
`pgx/v5` driver so generated code uses native types rather than
`database/sql` conversions.

The generated code is committed. This keeps sqlc out of the build closure, and
a `sqlc diff` check in the gate catches the obvious failure mode of SQL and
generated code drifting apart.

Alternative considered: writing queries by hand. Rejected because the schema
is about to grow fast and hand-written scanning is where silent column
mismatches live.

### Migrations run through goose as a library, holding an advisory lock

goose is embedded as a library rather than invoked as a CLI, with
`//go:embed` over `db/migrations`. The migration run takes a PostgreSQL
session advisory lock, so two instances starting together serialise: one
applies, the other waits and then finds nothing to do.

A binary older than the database is a fatal start-up error rather than a
warning. The alternative, serving anyway, means a server writing rows it does
not understand into a schema it cannot see, which is the kind of failure that
is only discovered later and by a customer.

Migrations are SQL files, not Go. A Go migration can do anything, which makes
it impossible to reason about from the file alone.

### The frontend is embedded through a copied directory

`internal/webui` owns an `//go:embed all:dist` directive over a `dist`
directory inside its own package, because `go:embed` cannot reach outside the
package. The Nix build for the server copies the built frontend into that
directory before compiling. The directory is committed containing only a
placeholder, so `go build` and `go test` work in a checkout where the frontend
was never built, and the server then serves a short development notice instead
of the application.

Alternatives considered: serving assets from a directory on disk, which breaks
the one-binary deployment; and a Go generator that writes a source file, which
makes the diff unreadable and the build order worse.

### The gate grows four checks

| Check         | What it proves                                               |
| ------------- | ------------------------------------------------------------ |
| `frontend`    | The frontend builds, type-checks and passes its unit tests    |
| `sqlc`        | The committed generated code matches the SQL it came from     |
| `integration` | Migrations run from empty, and the server serves, against a real PostgreSQL |
| `nixos-vm`    | Deferred to the deployment epic, named here so it is not forgotten |

`integration` starts a throwaway PostgreSQL inside the Nix build, on a unix
socket in the build directory. This works in the sandbox because nothing
leaves the machine. Tests that genuinely need a browser or a network wait for
the end-to-end epic and run as a NixOS VM test.

### Configuration is environment variables, with a `_FILE` escape

Every value is an environment variable. Any value may instead be supplied as
`<NAME>_FILE` pointing at a readable file, which is how the NixOS module keeps
the database password and the master key out of the Nix store. The loader
rejects both forms being set for the same value rather than picking one.

### Handlers take an actor from the request context

No handler reads a session or compares a role. A handler receives an actor
from the request context and, once the tenancy epic lands, asks
`internal/policy.Can`. This change installs the context plumbing and a single
anonymous actor, so inserting real identity later is an authentication change
rather than a rewrite of every handler.

This is the only one of the four seams in `docs/ARCHITECTURE.md` that this
change touches, and it touches it by leaving a hole of the right shape.

### Logging is `log/slog` in JSON

Structured from the start, with the request id as a standard attribute. A
human-readable handler is selected when the output is a terminal, so the dev
shell stays readable.

## Risks / Trade-offs

- **The first gate run gets much slower.** Building the Node dependency tree
  once is unavoidable. Mitigation: `npmDepsHash` is pinned, so it is built
  once per dependency change and cached afterwards.
- **A pinned `npmDepsHash` is a manual step.** Forgetting to update it after
  changing `package.json` produces a confusing failure. Mitigation: the
  failure message in the task list says exactly which command regenerates it,
  and the proposal records it as a known cost.
- **A PostgreSQL inside a Nix build is unusual.** It can fail on a machine
  whose sandbox forbids unix sockets in the build directory. Mitigation: the
  integration check is a separate check, so a failure points straight at the
  cause rather than at the whole gate.
- **Committed generated code gets stale.** Mitigation: the `sqlc` check fails
  the gate rather than leaving it to review.
- **The advisory lock serialises start-up.** Two instances starting together
  means one waits for the other's migrations. Mitigation: this is correct
  behaviour and the wait is bounded by a timeout that is itself a fatal error
  when exceeded.
- **An embedded placeholder can ship by accident.** A binary built without the
  frontend serves a development notice rather than the app. Mitigation: the
  notice says plainly that the frontend was not built, and the Nix package
  always builds it.

## Migration Plan

There is nothing deployed, so there is nothing to migrate. The first
deployment runs the binary against an empty database and the migrations create
the schema. Rollback is deleting the database.

From this change onwards, a rollback means deploying the previous binary. A
binary older than the database refuses to start, so a rollback across a
migration requires restoring the database as well. This is stated here so the
deployment epic can document it.

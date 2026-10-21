## 1. Configuration

- [x] 1.1 Extend `internal/app.Config` with the listen address, base URL, state
  directory, database URL, shutdown timeout, log format and the git and hugo
  binary paths; verify `TestLoadConfigAppliesDefaults` covers every field.
- [x] 1.2 Add `<NAME>_FILE` resolution that reads a value from a file, trims a
  single trailing newline, and errors when both forms are set; verify with
  `TestLoadConfigReadsSecretFromFile` and
  `TestLoadConfigRejectsBothValueAndFile` using `t.TempDir`.
- [x] 1.3 Make the config error name the missing variable and carry no value;
  verify with `TestLoadConfigErrorNamesVariableNotValue`.

## 2. HTTP foundation

- [x] 2.1 Add `internal/httpapi` with a `Router` built on `net/http.ServeMux`,
  the JSON error shape `{"error":{"code","message"}}`, and helpers for writing
  JSON responses; verify with `TestWriteErrorShape`.
- [x] 2.2 Add the 404 and 405 handlers for the `/api` prefix returning the JSON
  error shape; verify with `TestUnknownAPIPathReturnsJSON404`.
- [x] 2.3 Add request id middleware reading `X-Request-Id` or generating one,
  returning it in the response header and putting it in the context; verify
  with `TestRequestIDIsGenerated` and `TestRequestIDIsReused`.
- [x] 2.4 Add `log/slog` request logging with the request id, method, path,
  status and duration, choosing JSON or text by the configured log format;
  verify with `TestRequestLogCarriesRequestID` against a capturing handler.
- [x] 2.5 Add panic recovery returning a generic JSON 500 and logging the panic
  with its stack and request id; verify with `TestPanicIsRecoveredAndLogged`
  that the next request still succeeds.
- [x] 2.6 Add JSON body decoding that rejects unknown fields and oversized
  bodies with a 400 carrying no echo of the body; verify with
  `TestDecodeRejectsMalformedBody` and `TestDecodeRejectsUnknownField`.
- [x] 2.7 Add an `Actor` type and context plumbing with a single anonymous
  actor, so handlers read the actor from the context and never a session;
  verify with `TestHandlerReadsActorFromContext`.

## 3. Database

- [x] 3.1 Add `pgx/v5` and `pgxpool`, and an `internal/store.Open` that builds
  a pool from the config and pings it; verify with an integration test
  `TestOpenPingsDatabase` guarded by the integration build tag.
- [x] 3.2 Add `db/migrations/00001_init.sql` creating the `jobs` table and
  nothing else, since no epic yet owns a domain table; verify the migration
  applies and rolls back with `TestMigrateUpDown`.
- [x] 3.3 Embed `db/migrations` with `//go:embed` and add
  `internal/store.Migrate` using goose as a library; verify with
  `TestMigrateFromEmptyDatabase` that every migration applies in order.
- [x] 3.4 Take a PostgreSQL session advisory lock around the migration run and
  release it on both paths; verify with `TestConcurrentMigrateIsSafe` that two
  concurrent migrate calls both succeed and each migration applies once.
- [x] 3.5 Make a database at a version this binary does not know a fatal
  start-up error naming the condition; verify with
  `TestMigrateRefusesNewerDatabase`.
- [x] 3.6 Add `sqlc.yaml` and `db/queries/jobs.sql`, generate into
  `internal/store`, and commit the generated code; verify `sqlc diff` reports
  no difference.

## 4. Frontend

- [x] 4.1 Create `web/` with Vite, React, TypeScript and Tailwind v4 via
  `@tailwindcss/vite`, plus Vitest; verify `npm run build`, `npm run
  typecheck` and `npm run test` all pass in the dev shell.
- [x] 4.2 Add a placeholder application that calls `/api/health` and renders
  the result, so the proxy, the API shape and the embed are all exercised;
  verify with a Vitest test `renders the health result` against a mocked
  fetch.
- [x] 4.3 Configure the dev server to proxy `/api` to the Go server, with the
  target taken from an environment variable defaulting to the Go server's
  default listen address; verify by running both and loading the page.
- [x] 4.4 Add `internal/webui` with `//go:embed all:dist`, a committed
  placeholder in `dist`, and a handler that serves hashed assets with an
  immutable cache header, the application document with no-cache, and the
  application document for unknown non-API paths; verify with
  `TestAssetCacheHeaders`, `TestDocumentIsNotCached` and
  `TestUnknownPathServesDocument`.
- [x] 4.5 Make the handler serve a plain development notice when `dist` holds
  only the placeholder; verify with `TestUnbuiltFrontendServesNotice`.

## 5. Server lifecycle

- [x] 5.1 Rework `internal/app.Run` to load config, open the pool, migrate,
  build the router and only then open the listener; verify with
  `TestRunDoesNotListenWhenMigrationFails` that no port is bound on failure.
- [x] 5.2 Add `/api/health` reporting server version, schema version and
  database reachability, with 200 when reachable and 503 when not, and no
  connection string in either body; verify with `TestHealthOK` and
  `TestHealthReportsUnreachableDatabase`.
- [x] 5.3 Add graceful shutdown on interrupt and terminate: stop accepting
  connections, drain in-flight requests within the shutdown timeout, exit
  zero; verify with `TestShutdownDrainsInFlightRequest`.
- [x] 5.4 Exit non-zero and log that shutdown was forced when a request
  outlives the timeout; verify with `TestShutdownTimeoutIsForced`.

## 6. The gate

- [x] 6.1 Add the `web` package to `flake.nix` with `buildNpmPackage` and a
  pinned `npmDepsHash`, and make the server package copy its output into
  `internal/webui/dist` before compiling; verify `nix build .#chatcmsd`
  produces a binary that serves the application rather than the notice.
- [x] 6.2 Add the `frontend` check running the frontend type check, unit tests
  and build; verify it fails when a deliberate type error is introduced.
- [x] 6.3 Add the `sqlc` check running `sqlc diff`; verify it fails when a
  query is edited without regenerating.
- [x] 6.4 Add the `integration` check that starts a throwaway PostgreSQL on a
  unix socket inside the build directory and runs the integration-tagged
  tests; verify it fails when a migration is broken.
- [x] 6.5 Set `vendorHash` now that the module has dependencies, and confirm
  `nix build .#chatcmsd` works with no network beyond the fixed-output
  fetches; verify with `nix build --offline` after one warm build.
- [x] 6.6 Teach `scripts/coverage-gate.sh` to exclude generated files by
  filename pattern, and test-support packages by path, and document both in the
  script; verify the reported overall figure changes accordingly and the script
  still fails below the threshold. Excluding by filename rather than by package
  keeps the hand-written code in `internal/store` measured.
- [x] 6.7 Add `internal/httpapi` to the core package list in
  `scripts/coverage-gate.sh` once it exists; verify the gate enforces 80% on
  it.

## 7. Documentation and shipping

- [x] 7.1 Update `AGENTS.md` with the new commands (`sqlc generate`, the
  frontend scripts, how to regenerate `npmDepsHash`) and the new repository
  paths; verify the commands in it run as written.
- [x] 7.2 Add a bullet to `CHANGELOG.md` under `## [Unreleased]` and
  `### Added` describing the server, the schema and the frontend from a user's
  perspective.
- [x] 7.3 Run `nix flake check` and confirm every check passes, including the
  three new ones.

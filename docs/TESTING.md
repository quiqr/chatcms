# Testing strategy

The PoC has to prove it works, not look like it works. Every epic lands with
tests, and `nix flake check` is the one command that decides whether a change
ships.

## Layers

| Layer       | Where                      | Runs against                          | Gate              |
| ----------- | -------------------------- | ------------------------------------- | ----------------- |
| Unit        | `internal/*/..._test.go`   | Pure Go, no I/O                       | Coverage gate      |
| Golden      | `test/fixtures/`           | Real Quiqr models and real site files  | Byte equality      |
| Integration | `internal/*/..._test.go`   | Real Postgres, real git repos on disk  | Must pass          |
| End to end  | `test/e2e/`                | The built binary and a real browser    | Must pass          |

## Golden tests: the backwards-compatibility contract

This is the most important test category in the project. The new app must read
existing Quiqr models unchanged, and saving through a form must produce the
same front matter and Markdown as the old Quiqr desktop app.

How it works:

1. A fixture is a site directory under `test/fixtures/sites/<name>/`, with its
   `quiqr/model/` and a handful of content files.
2. For each fixture the golden file records, per content file, the exact bytes
   the old Quiqr produces after a load-and-save round trip with no edits.
3. The test loads the file through `internal/model` and `internal/forms`,
   writes it back, and compares bytes. Any difference fails.
4. A second variant applies a known edit per field type and compares against a
   golden for that edit, so key order, quoting style, date formats, list
   indentation and body separators are all pinned.

The reference installation is the old Quiqr desktop app at
`../quiqr-desktop`. Goldens are generated from it, never hand-written. When a
golden has to change, the commit must say which old-Quiqr behaviour it matches
and why.

Field types that need their own fixture coverage: text, date, select, image,
textarea, markdown body, list of items, nested and collapsed groups, boolean,
number, and any remaining type the reference app supports.

## Unit tests

The four core packages carry the 80% coverage floor because their failures are
silent and expensive:

- `internal/model` parses the Quiqr model. A misparse silently drops fields.
- `internal/forms` does the round trip. A bug here corrupts customer content.
- `internal/policy` is the only authorisation decision. A bug here is a breach.
- `internal/gitsync` owns the repository. A bug here loses work.

Policy gets a table test over the full cross product of role, action and
resource, with the expected answer written out explicitly rather than computed,
so the table itself is the specification.

## Integration tests

No mocks for Postgres and no mocks for git.

- Postgres comes from `scripts/dev-db.sh`, or from a per-test temporary cluster
  in CI. Each test gets its own schema or its own database and migrations run
  from empty, which also proves the migration path.
- Git remotes are bare repositories in a temp directory. Push, fetch, a commit
  landing on the remote behind the app's back, a conflicting change, and a
  force-pushed remote are all reproducible this way.
- Hugo is the pinned binary from the flake, so preview and build tests use the
  same version production does.

## End-to-end tests

Playwright, in `test/e2e/`, driving the built binary against a real database,
a real bare git remote and real fixture sites. One spec per acceptance test
from `docs/BRIEFING-M01.md` section 6:

| # | Acceptance test                                                         |
| - | ----------------------------------------------------------------------- |
| 1 | Admin creates an organisation and invites a customer as editor            |
| 2 | Import a Hugo repository with a Quiqr model using a deploy key            |
| 3 | All model forms render; saving matches old Quiqr front matter             |
| 4 | Create, edit and delete a page, and upload an image                      |
| 5 | Live preview shows a change within a few seconds                         |
| 6 | A commit pushed straight to the remote shows up in the app automatically  |
| 7 | Publishing shows what changed, then pushes                               |
| 8 | Roll back to an earlier version                                          |
| 9 | Every flow works at 390 px                                              |

Every spec runs twice, at 1440 px and at 390 px. The phone pass also asserts no
horizontal page scrolling and no touch target under 44 px, which the design
calls for.

Two further checks that are not screen flows:

- The app builds and deploys through the flake and the NixOS module, tested
  with a NixOS VM test that boots the module, waits for the service and fetches
  the sign-in page.
- Migrations run cleanly from an empty database, asserted by the integration
  harness on every run.

## What the gate runs

`nix flake check` is the contract. As epics land, each adds its check:

| Check          | Added by                        |
| -------------- | ------------------------------- |
| `go-build`     | Present                         |
| `go-test`      | Present, includes the coverage gate |
| `frontend`     | The foundation epic             |
| `golden`       | The form round-trip epic        |
| `e2e`          | The acceptance suite epic       |
| `nixos-vm`     | The deployment epic             |

A check that cannot run in the Nix sandbox, because it needs a browser or a
network, runs in a NixOS VM test instead. Nothing is excused from the gate for
being awkward.

## Golden files

Two kinds of golden file exist, and they are regenerated the same way.

| Golden                           | Pins                                             |
| -------------------------------- | ------------------------------------------------ |
| `test/fixtures/models/*/model.golden.json` | The JSON shape of a parsed Quiqr model and its navigation |
| `test/fixtures/sites/*/`                   | The bytes the old Quiqr writes for a content file |

Regenerate with:

```bash
CHATCMS_UPDATE_GOLDEN=1 go test ./...
```

That variable is never set during a normal run, so a changed golden always
fails the gate first. Read the diff before committing it: a golden that changes
without a reason in the commit message is a behaviour change nobody noticed.

## Conventions

- Table-driven tests, with the case name as the map key.
- `t.Setenv` and `t.TempDir` rather than hand-rolled cleanup.
- No sleeps. Poll with a deadline, or expose a hook the test can wait on.
- A test that needs a fixture site creates it from `test/fixtures/`, never from
  a live customer repository.
- Failures print what was expected and what happened, in that order.

# Quiqr ChatCMS

Quiqr ChatCMS (chatcms.eu) is a hosted, multi-tenant, Nix-powered headless CMS
for Markdown-based static sites, Hugo first. Its core is generating editing
forms from the declarative Quiqr model (YAML), so non-technical people can
safely edit site content. Content lives in git, and git is hidden from content
creators.

Two audiences:

- **Implementers** (template developers, site implementers) create projects,
  import repositories, configure publishing and manage users.
- **Content creators** (customers) edit pages through forms, preview, and press
  Publish. They never see git terminology.

Milestone 01 is a fully usable MVP that can be deployed for a real customer
whose Hugo site currently lives in the old Quiqr desktop app. The briefing that
defines it is `docs/BRIEFING-M01.md`; the acceptance tests in its section 6 are
the definition of done. `docs/KICKOFF.md` describes the session loop.

## Locked decisions

Do not revisit these without asking.

| Area       | Decision                                                         |
| ---------- | ---------------------------------------------------------------- |
| Backend    | Go                                                                |
| Frontend   | React + TypeScript + Vite + Tailwind v4 + shadcn/ui, responsive   |
| Database   | PostgreSQL, queries via sqlc, migrations via goose                |
| Migrations | Embedded in the binary, run on start-up                           |
| Git        | The git CLI called from Go, never go-git, version pinned via Nix  |
| Packaging  | Nix flake plus a NixOS module; Nix-only deployment is fine        |
| VCS        | jj, colocated with git; remote `git@github.com:quiqr/chatcms.git` |

The database holds users, organisations, memberships, roles, projects,
credentials and job state. Content stays in git.

## Repository layout

| Path                | What lives there                                          |
| ------------------- | ---------------------------------------------------------- |
| `cmd/chatcmsd/`     | The server entry point, nothing but a call into `internal/app` |
| `internal/`         | All Go code; see `docs/ARCHITECTURE.md` for the packages   |
| `web/`              | The React frontend, built by Nix and embedded in the binary |
| `db/migrations/`    | goose migrations, embedded into the binary                 |
| `db/queries/`       | sqlc query files                                           |
| `nix/`              | The NixOS module and other Nix glue                        |
| `scripts/`          | Shipping, coverage gate, dev database                      |
| `docs/design/`      | UI design hand-off, reference material, never imported      |
| `docs/`             | Briefing, architecture, testing strategy                   |
| `openspec/`         | Change proposals and specs                                 |
| `.beans/`           | Milestones and epics                                       |
| `test/e2e/`         | Playwright acceptance tests                                |
| `test/fixtures/`    | Quiqr model fixtures and golden files                      |

## Commands

```bash
nix develop                        # dev shell with Go, Node, Postgres, sqlc, goose, Hugo
nix flake check                    # the gate: build, vet, tests, coverage, frontend, sqlc
nix build .#chatcmsd               # build the server with the frontend embedded
nix build .#web                    # build the frontend on its own

go test ./...                      # fast unit loop, no database needed
go test -tags=integration ./...    # adds the tests that need PostgreSQL
bash scripts/dev.sh                # everything: database, key, seed, server on :8080
bash scripts/dev.sh fresh          # the same, throwing the database and state away first
bash scripts/dev-db.sh up          # local Postgres in ./.pgdata, prints the URL
bash scripts/dev-db.sh reset       # throw the database away and start over

sqlc generate                      # regenerate query code after editing db/queries
sqlc diff                          # what the gate checks: is the generated code stale
goose -dir db/migrations postgres "$CHATCMS_DATABASE_URL" up

cd web && npm run dev              # frontend dev server, proxies /api to the Go server
cd web && npm run typecheck        # what the frontend check runs
cd web && npm run test             # Vitest

bash scripts/ship-change.sh <change> "<commit subject>"
beans list --ready                 # the next epic that is not blocked
beans roadmap                      # milestones and epics as Markdown
openspec list                      # active change proposals
```

After changing `web/package.json`, the pinned `npmDepsHash` in `flake.nix` is
stale and the build fails. Regenerate it with:

```bash
nix develop --command prefetch-npm-deps web/package-lock.json
```

The server embeds `internal/webui/dist`, which holds only a placeholder in a
checkout. `nix build .#chatcmsd` fills it from the frontend build; a plain
`go build` leaves it empty and the server then serves a short notice saying the
frontend was not built.

UI work follows `docs/design/DESIGN.md`; reference screens are in
`docs/design/reference/`.

## How to work

The `/mip:1shotpoc` workflow and its gates are leading. Where this file
conflicts with that workflow on process, the workflow wins.

1. Pick the next epic from beans (`beans list --json --ready`).
2. Write one OpenSpec change proposal per epic, linking back to the bean.
3. Implement every task in `tasks.md`, with tests, checking each one off.
4. Ship with `scripts/ship-change.sh`, which gates on `nix flake check` before
   it archives, commits or pushes anything.
5. Close the bean with a `## Summary of Changes` section.

Work autonomously. Stop for a question only when a decision is truly blocking
and not covered by the briefing, the design, the reference app or this file.
Log assumptions in the OpenSpec proposal instead of asking.

### The gate

`nix flake check` must stay green. It runs these checks:

| Check         | What it runs                                                     |
| ------------- | ---------------------------------------------------------------- |
| `go-build`    | The server package, frontend included                             |
| `go-test`     | `go vet` and the unit tests with the race detector                |
| `integration` | The same plus the integration-tagged tests against a real PostgreSQL, then the coverage gate |
| `frontend`    | The frontend type check, Vitest and the production build          |
| `sqlc`        | `sqlc diff`, so committed generated code cannot drift from its SQL |

Before the gate runs, `scripts/ship-change.sh` also runs
`scripts/check-specs.sh <change>`, which fails when a change's proposal
declares a capability it has no spec file for. `nix flake check` cannot see a
spec that was never written, and `openspec validate` is satisfied by one delta,
so without this a change can ship with a capability it only promised.

`scripts/coverage-gate.sh` requires at least 70% statement coverage overall and
at least 80% on each core package: `internal/model`, `internal/forms`,
`internal/policy`, `internal/gitsync` and `internal/httpapi`. Generated files
and test-support packages are excluded. A core package that does not exist yet
is reported and skipped, so the gate tightens as those packages appear. Never
bypass the gate; fix the code.

Each epic that adds a buildable artifact must also add its check to
`flake.nix`, so the e2e suite and the NixOS module test end up behind the same
single command.

### Version control

jj, colocated with git. Commit after every archived OpenSpec change, as Pim
Snel, with no self-promotion: no `Co-authored-by: Claude`, no "Generated with
Claude Code", no Anthropic attribution of any kind.

## Testing policy

See `docs/TESTING.md` for the full strategy. The short version:

- **Golden tests for backwards compatibility.** Reading an existing Quiqr model
  and saving through a form must produce byte-identical front matter and
  Markdown to the old Quiqr. Fixtures live in `test/fixtures/`, goldens next to
  them. This is the single most important test category in the project.
- **Unit tests** for the model parser, the form round-trip, the policy function
  and the git sync engine. These are the core packages behind the 80% gate.
- **Integration tests** against a real Postgres and a real git remote on disk,
  not mocks.
- **End-to-end tests** in Playwright covering every acceptance test from the
  briefing, run at 1440 px and at 390 px.

## Copy rules

Editor-facing UI never uses git words: commit, branch, push, pull, merge,
repository, deploy, build log, HEAD, SHA, conflict. Use saved, changes,
publish, version, go back to, "edited by two people". Technical detail is
allowed for owners and implementers, in clearly secondary places. Full rules in
`docs/design/DESIGN.md` section 7.

## Beans

When I refer to issues like chatcms-rn3b checkout the task
in @.beans/chatcms-rn3b-*.md

In this project we will use these tasks as epics for making openspec proposals.

WHEN you create a proposal at a link to this task in the proposal.md.
WHEN a bean is used to create an proposal change the status to "in-progress"
WHEN a proposal is archived add the link to the archived proposal in the frontmatter of this task like this:

```
openspec-link: openspec/changes/archive/....
```

You are allowed to update these statuses in the task frontmatter:

- in-progress
- todo
- draft
- completed
- scrapped

When making changes you are allowed to update the date/time in `updated_at` in the task frontmatter

Besides updating status and openspec-link, you are NOT ALLOWED to modify the contents of the task file.

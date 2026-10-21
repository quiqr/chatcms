# Quiqr ChatCMS

A hosted, multi-tenant, Nix-powered headless CMS for Markdown-based static
sites, Hugo first. Editing forms are generated from the declarative Quiqr model
(YAML), so non-technical people can safely edit site content. Content lives in
git, and git stays hidden from content creators.

Status: milestone 01, building the MVP. Not yet usable.

## Getting started

```bash
nix develop                  # Go, Node, Postgres, sqlc, goose, Hugo, Playwright
bash scripts/dev-db.sh up    # local Postgres, prints CHATCMS_DATABASE_URL
nix flake check              # build, vet, tests, coverage gate
```

## Documentation

| File                     | What it covers                                   |
| ------------------------ | ------------------------------------------------ |
| `AGENTS.md`              | How to work here; also read as `CLAUDE.md`       |
| `docs/KICKOFF.md`        | How to start and run an autonomous build session |
| `docs/BRIEFING-M01.md`   | The milestone 01 brief and its acceptance tests  |
| `docs/ARCHITECTURE.md`   | Package layout and the seams that must stay thin |
| `docs/TESTING.md`        | Test layers, golden tests, the gate              |
| `docs/design/DESIGN.md`  | The UI spec and reference screens                |
| `openspec/`              | Change proposals and specs                       |
| `.beans/`                | Milestones and epics                             |

## Licence

See `LICENSE`.

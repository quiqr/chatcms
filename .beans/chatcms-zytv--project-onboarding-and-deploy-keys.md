---
# chatcms-zytv
title: Project onboarding and deploy keys
status: in-progress
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T18:02:39Z
parent: chatcms-8vem
blocked_by:
    - chatcms-gw8g
    - chatcms-1pwg
---

## Purpose

Get a real site into the system: import an existing Hugo repository with a
Quiqr model, generate a deploy key, and configure publishing.

## Scope

- `internal/secrets`: encryption at rest for per-project SSH deploy keys, with
  the master key read from `CHATCMS_SECRET_KEY_FILE`, never from the Nix store.
- Generate an SSH deploy key per project, show the public half, and offer
  "Test connection again".
- Detect what was imported: Hugo site, Quiqr model found or not, with a clear
  result either way.
- Three import sources: a git repository URL, a site exported from the old
  Quiqr desktop app, and a Quiqr template.
- Publishing configuration: push to main or to a separate branch defaulting to
  `quiqr/publish`, and push either source files or built HTML.
- The onboarding screen from DESIGN.md section 5.3, with its step list and
  three section cards.

## Done when

- Importing the customer repository with a deploy key works end to end, which
  is acceptance test 2.
- The private key is never readable from the API, never logged, and never in
  the Nix store.
- A repository without a Quiqr model gives a useful message rather than a
  broken project.
- Importing a desktop-app export produces the same project as importing its
  repository.

## References

- `docs/BRIEFING-M01.md` section 4.3.
- `docs/design/DESIGN.md` section 5.3.

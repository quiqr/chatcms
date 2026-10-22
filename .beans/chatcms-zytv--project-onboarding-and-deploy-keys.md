---
# chatcms-zytv
title: Project onboarding and deploy keys
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-22-add-project-onboarding
updated_at: 2026-10-22T18:24:49Z
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

## Summary of Changes

Shipped as OpenSpec change `add-project-onboarding`, archived at `openspec/changes/archive/2026-10-22-add-project-onboarding`. Three capabilities built, `tenancy/secrets`, `onboarding/import` and `onboarding/publishing-target`, with deltas on `tenancy/organisations` and `sync/repository`.

Acceptance test 2 is proven end to end through the API: set a project up, make its key, import the customer site from a remote, see the Hugo configuration and the Quiqr model found with their page and collection counts, test the connection, and record where it publishes. A second test calls every endpoint that mentions a project and checks that no private key appears in any answer.

Secrets are sealed with AES-256-GCM under one master key read from a file, so a copy of the database is worth nothing and the key is never an environment variable. Deploy keys are Ed25519 generated in process rather than by running ssh-keygen, so no private key is ever written to a temporary file on the way. Each project has its own key; git reaches a remote with it from an owner-only directory made at start-up and removed at shutdown.

Importing has one path with three doors: a repository is cloned, an export is copied into a new working copy and pushed, a template is a repository ChatCMS knows. A test imports the same site both ways and compares every file, which is what makes the three doors one path rather than three.

A site with no model is reported as needing one rather than refused, because implementers import first and write the model second. What was looked for is named, so a site with an unusual layout can be understood rather than guessed at.

The onboarding screen is built with its step list and three sections.

One change to a shipped package: a save that names no file now saves everything the working copy holds, which is what importing is. Added to the `sync/repository` spec in this change.

Coverage: `internal/onboarding` 86.5%, `internal/secrets` 86.4%, overall 85.8%. `cmd/seed`, a development tool that fills a database, is excluded from the gate the way the test-support package is.

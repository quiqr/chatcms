---
# chatcms-va5c
title: End-to-end acceptance suite
status: todo
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:32Z
parent: chatcms-8vem
blocked_by:
    - chatcms-5uu2
    - chatcms-0aeg
    - chatcms-zytv
    - chatcms-ihey
    - chatcms-ac9e
    - chatcms-tp3b
---

## Purpose

Prove the PoC works. Every acceptance test from the briefing, automated, run
against the built binary, a real database, a real remote and real fixture
sites.

## Scope

- `test/e2e/` with Playwright, one spec per acceptance test from
  `docs/BRIEFING-M01.md` section 6.
- Every spec runs twice, at 1440 px and at 390 px. The phone pass also asserts
  no horizontal page scrolling and no touch target under 44 px.
- A reproducible harness: build the binary, start Postgres, create a bare
  remote, seed fixture sites, run, tear down.
- The `e2e` check in `flake.nix`, as a NixOS VM test if the browser cannot run
  in the sandbox.
- A run against the real customer site, scripted and repeatable, documented in
  the proposal.
- Screenshot comparison against `docs/design/reference/screenshots/` for each
  screen, with differences listed rather than silently accepted.

## Done when

- All nine acceptance tests pass automatically in `nix flake check`.
- The suite is green three times in a row with no flake, and uses no sleeps.
- A deliberately broken behaviour makes the matching spec fail.

## References

- `docs/BRIEFING-M01.md` section 6.
- `docs/TESTING.md`, the end-to-end section.

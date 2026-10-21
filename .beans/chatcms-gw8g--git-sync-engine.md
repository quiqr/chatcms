---
# chatcms-gw8g
title: Git sync engine
status: todo
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:31Z
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

The repository engine. This has to be rock solid: it owns the customer's work.
Everything that touches a repository goes through one interface so Jujutsu can
replace git later.

## Scope

- `internal/gitsync`: the `Repo` interface and its git CLI implementation.
  Clone, working copy management, auto-commit on save, fetch, merge or rebase,
  push, and state reporting. The git binary comes from `CHATCMS_GIT_BIN`.
- No `exec.Command("git", ...)` anywhere outside this package, enforced by a
  test or lint over the tree.
- `internal/jobs`: a Postgres-backed queue and worker for background sync, with
  retries, backoff, and visibility of job state.
- Detect and pull new commits from the remote on a schedule and on demand.
- Conflict handling: detect, surface, and offer the resolutions the design
  describes. Same-field edits by two people produce the inline question, not a
  merge marker.
- A projection of git state into Postgres for the unpublished change count,
  rebuildable from git alone.
- Integration tests against bare repositories on disk: push, fetch, a remote
  commit landing behind the app's back, a conflicting change, and a
  force-pushed remote.

## Done when

- `internal/gitsync` is at or above 80% statement coverage.
- A commit pushed straight to the remote appears in the app automatically,
  which is acceptance test 6.
- No test uses a git mock.
- Killing the process mid-sync leaves a recoverable state, proven by test.

## References

- `docs/BRIEFING-M01.md` section 4.7.
- `docs/ARCHITECTURE.md`, the seams section.

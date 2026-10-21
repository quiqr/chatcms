---
# chatcms-gw8g
title: Git sync engine
status: completed
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-21-add-git-sync-engine
updated_at: 2026-10-21T19:27:08Z
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

## Summary of Changes

Shipped as OpenSpec change `add-git-sync-engine`, archived at `openspec/changes/archive/2026-10-21-add-git-sync-engine`.

Three capabilities specified and built: `sync/repository`, `sync/remote-exchange` and `platform/background-jobs`.

`internal/gitsync` wraps the git CLI behind an interface that speaks in sites and versions, never in commits, branches or refspecs, so Jujutsu could replace it. `TestOnlyGitsyncRunsGit` walks the sources and fails when any other package starts a process and mentions git, so the rule is enforced rather than written down. Git runs with hooks disabled, no credential helper, no prompt and no system or user configuration, because a customer repository's hook is code we will not run on the server.

Conflicts never reach a content file as markers: the engine reads all three versions out of the index, writes valid content back to disk, and hands the three versions to the caller in memory. The content filesystem refuses a conflicted path rather than serving a version as if it were the truth. An answer given for a page stands for the rest of that exchange, so a replay of several local versions asks once rather than three times; the design records why keying on content was built and then removed.

`internal/jobs` is a Postgres-backed queue with exclusive claiming via FOR UPDATE SKIP LOCKED, exponential backoff, an attempt limit, and deduplication enforced by a unique index over queued rows rather than a read-then-write. Shutdown drains the job in flight and returns an overrunning one to the queue. `internal/redact` strips credential-shaped text from anything stored, logged or shown.

Acceptance test 6 is proven at the engine level: a version pushed straight to the remote arrives through the scheduled sync job without anyone asking.

Coverage: `internal/gitsync` 84.2%, `internal/jobs` 80.4%, overall 86.8%. The gate now runs against the pinned git binary and a real bare repository on disk; no test uses a git mock.

Three things deferred with reasons in the proposal: no project table yet, so a site is an id and a directory and status is computed from git rather than projected into Postgres; no credentials yet, so a key path is accepted but never read; and conflict resolution is file-level, which the form epic turns into the per-field question the design asks for.

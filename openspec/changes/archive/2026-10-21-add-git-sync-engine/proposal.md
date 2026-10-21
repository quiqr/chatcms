## Why

Everything a content creator does ends up in a git repository they never see.
If that layer loses work, reorders it, or silently diverges from the remote,
the product has failed in the one way a CMS cannot recover from. Epic
`chatcms-gw8g`.

It serves acceptance test 6 from `docs/BRIEFING-M01.md` section 6, "a commit
pushed directly to the remote shows up in Quiqr automatically", and it is the
precondition for acceptance tests 2, 4, 7 and 8.

## What Changes

- **One place calls git.** `internal/gitsync` wraps the git CLI behind an
  interface that speaks in sites and versions rather than in commits and
  branches. Nothing else in the tree runs git, enforced by a test over the
  source.
- **A working copy per site.** Cloning, opening, and reporting on a working
  copy, with the git binary taken from configuration rather than from `PATH`.
- **Saving makes a version.** Writing content produces a commit attributed to
  the person who edited, with the system as committer, so history is readable
  by a human and attributable to a user.
- **Sync is a background job.** `internal/jobs` is a PostgreSQL-backed queue
  and worker with retries, backoff and visible state, so a slow or failing
  remote never blocks an editor.
- **Remote changes arrive on their own.** Fetching on a schedule and on
  demand, and integrating remote commits into the working copy without the
  editor asking.
- **Conflicts are data, not merge markers.** When the same file changed on
  both sides, the engine reports the conflict with all three versions of the
  file, and accepts a resolution. A merge marker never reaches a content file
  an editor can open.
- **Status answers the questions the interface asks.** Whether everything is
  saved, whether a sync is running, whether the remote moved, and how many
  changes are unpublished, all computed from git rather than from a flag the
  application maintains.

## Capabilities

### New Capabilities

- `sync/repository`: how a site's working copy is created, read, written and
  reported on, and what a version is.
- `sync/remote-exchange`: how local versions reach the remote and remote
  versions reach the working copy, including conflicts and their resolution.
- `platform/background-jobs`: how work that must not block a request is
  queued, claimed, retried and observed.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/gitsync`, `internal/jobs`.
- **Changed code**: `internal/app` gains the worker in its lifecycle, so
  shutdown now drains in-flight jobs as well as in-flight requests.
- **Dependencies added**: none. The git CLI is called as a subprocess, which
  is a locked decision in `AGENTS.md`.
- **Schema**: one migration adds a subject to the jobs table with a unique
  index over queued work, so asking twice for the same sync cannot queue it
  twice. The design records why the database enforces this rather than Go.
- **Seams**: this change implements two of the four seams named in
  `docs/ARCHITECTURE.md`. `gitsync.Repo` is the one that must stay thin so
  Jujutsu can replace git; the design records what that costs.
- **Coverage**: `internal/gitsync` becomes a core package carrying the 80%
  floor.
- **Gate**: the integration check grows a real git remote on disk, so the
  check now needs the git binary as well as PostgreSQL.

## Assumptions

- **No project table yet.** A site is identified by an opaque id and a
  directory. Persisting a sync projection per project needs projects, which
  the tenancy and onboarding epics own. Status is therefore computed from git
  on demand, which is the part that has to be rebuildable anyway; caching it
  in PostgreSQL is a later, purely additive change.
- **No credentials yet.** Reaching a remote over SSH needs a deploy key, which
  the onboarding epic owns. This change accepts a key path and configures git
  to use it, and is tested against local repositories on disk, which need no
  key. Nothing here reads or writes a key.
- **Rebase, not merge.** Local versions are replayed onto the remote rather
  than merged, so history stays a readable list of edits instead of a graph
  with merge commits an editor would see as noise. The design records what
  this costs when a conflict happens.
- **One writer per working copy.** Two requests editing the same site are
  serialised by a per-site lock inside the process. Two server instances
  editing the same site are out of scope for milestone 1 and are named in the
  design as a known limit.
- **Conflict resolution is file-level here.** The engine reports a conflicted
  file with its three versions and accepts replacement content. Turning that
  into the per-field question the design asks for needs the form layer, and
  belongs with the form epic. This change makes that possible and does not do
  it.

# Tasks

## 1. The interface

- [x] 1.1 `internal/runner`: `Task`, `Result`, `Runner`, with no host detail
- [x] 1.2 Validate a task: refuse absolute or climbing output paths before
      anything runs
- [x] 1.3 A table of cases both implementations must satisfy

## 2. The sandbox

- [x] 2.1 Probe isolation once, when the runner is made; refuse to run without it
- [x] 2.2 Lay out the workspace: inputs copied in, a writable working directory,
      a temporary area, tools read-only
- [x] 2.3 Network off unless the task asks
- [x] 2.4 Time limit, with the log kept and nothing left running
- [x] 2.5 Memory limit via `prlimit`
- [x] 2.6 Collect named outputs; a missing one is an error
- [x] 2.7 Remove the workspace whatever happened

## 3. Proving isolation

- [x] 3.1 Another project's working copy is unreadable
- [x] 3.2 The deploy key directory is unreadable
- [x] 3.3 The server's environment is not inherited
- [x] 3.4 The network is unreachable
- [x] 3.5 Writing outside the workspace fails, inside succeeds
- [x] 3.6 An unnamed file does not come back
- [x] 3.7 Cancelling leaves no process and no workspace

## 4. Building a site

- [x] 4.1 `internal/sitebuild`: build a site with the pinned Hugo through a runner
- [x] 4.2 Compare against Hugo run directly on the same site, file by file
- [x] 4.3 A site that fails to build reports what Hugo said

## 5. The record

- [x] 5.1 Migration: the builds table
- [x] 5.2 Queries and the store methods
- [x] 5.3 Bound the log, keeping the end and saying output was dropped
- [x] 5.4 Mark builds left running as interrupted on start-up
- [x] 5.5 List a project's builds, recent first

## 6. Shipping

- [x] 6.1 `bubblewrap` and `util-linux` in the dev shell, the checks and the
      NixOS module
- [x] 6.2 Settings for the limits, with defaults
- [x] 6.3 `docs/ARCHITECTURE.md`: the runner seam, and what replacing it costs
- [x] 6.4 The gate green, coverage held

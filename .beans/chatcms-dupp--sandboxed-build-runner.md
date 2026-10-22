---
# chatcms-dupp
title: Sandboxed build runner
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T21:40:00Z
openspec-link: openspec/changes/archive/2026-10-22-add-sandboxed-runner
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

Build the site in isolation, behind an interface, so the Nix sandbox used now
can be replaced by Firecracker micro VMs later without touching callers.

## Scope

- `internal/runner`: the `Runner` interface, speaking in inputs, a command and
  collected outputs rather than local paths. One implementation using a Nix
  sandbox.
- Resource limits, a timeout, and no network access during a build unless the
  workflow asks for it.
- Build output collection, log capture, and a build record in Postgres.
- Hugo builds run through the runner, with the pinned Hugo.

## Done when

- A build of the customer site produces the same output as running Hugo
  directly.
- A build that exceeds its timeout or its memory limit fails cleanly with a log
  the implementer can read.
- A second implementation can be added without changing any caller, argued in
  the proposal and checked by the interface shape.

## References

- `docs/BRIEFING-M01.md` section 4.9.
- `docs/ARCHITECTURE.md`, the seams section.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-22-add-sandboxed-runner`.

A build now runs inside bubblewrap with its own namespaces: it sees the files
it was given and the tools it was named, and nothing else. The isolation is
proven rather than asserted, by tests that put another project's working copy
and a deploy key on the machine and then try to read them from inside a build.

- `internal/runner` holds the interface and the bubblewrap implementation. A
  task is inputs, a command, limits and the paths it wants collected; a result
  is an exit code, a log, those outputs and which limit was reached. Nothing
  in either names a process, a mount or a path on this machine.
- A second implementation that isolates nothing is run against the same table
  of cases in the test package, so a property only bubblewrap can have cannot
  quietly become part of the interface.
- No network unless the task asks and the runner allows, no writing outside
  the workspace, a time limit, and a memory limit through `prlimit` on the
  data segment. An address-space limit was measured and rejected: a Go program
  needs more than a gigabyte of address space to print its own version, which
  makes any useful limit indistinguishable from a broken one.
- A runner that cannot isolate refuses to run anything. The server starts and
  says builds are unavailable rather than running them unprotected.
- `internal/sitebuild` runs the pinned Hugo through a runner. A test builds
  the site fixture both ways and compares every file.
- `internal/builds` records every build: the project, the reason, the start
  and finish, the state and a bounded log that keeps the end. Builds left
  running when the server stopped are marked interrupted at the next start.
- `bubblewrap` and `util-linux` are in the dev shell, the checks and the NixOS
  module, with settings for both tools and for the two limits.

Bubblewrap was checked in the Nix build sandbox before the design was settled,
so the gate exercises the real sandbox: 35 tests pass there and the one skip
is the check that sharing the network adds interfaces, which a Nix build's own
loopback-only network cannot show.

Nothing calls a build from an HTTP path yet. Workflows is the epic that gives
builds a reason to happen and a queue to happen in.

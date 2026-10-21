---
# chatcms-dupp
title: Sandboxed build runner
status: todo
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:31Z
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

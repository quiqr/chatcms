---
# chatcms-uwdh
title: Firecracker runner
status: draft
type: epic
priority: deferred
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
parent: chatcms-2udf
---

## Intent

Replace the Nix sandbox with Firecracker micro VMs for stronger isolation of
customer builds.

## Room left for it

`internal/runner.Runner` speaks in inputs, a command and collected outputs
rather than local paths, precisely so a VM implementation can be dropped in.
See `docs/ARCHITECTURE.md`.

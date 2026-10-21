---
# chatcms-2udf
title: 02 Assistant, isolation and scale
status: draft
type: milestone
priority: low
created_at: 2026-10-21T15:32:56Z
updated_at: 2026-10-21T15:32:56Z
---

## Goal

Everything milestone 01 deliberately left out, held here so it is visible
without blocking the MVP. Nothing in this milestone is specified yet; the
epics under it are placeholders that record the intent and the room left for
them in the architecture and the design.

## Why it exists now

Milestone 01 makes specific concessions on the assumption that these land
later: the runner interface exists so Firecracker can replace the Nix sandbox,
the policy function takes a resource argument so per-document permissions can
be added, the editor keeps a three-slot layout so the assistant panel fits, and
the auth identity source is an interface so OIDC can be added. If an epic here
is dropped, the matching concession can be simplified away.

## Source

`docs/BRIEFING-M01.md` section 5.

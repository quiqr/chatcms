---
# chatcms-8d4z
title: Sync status and conflict surfacing
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T20:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-sync-status
parent: chatcms-8vem
blocked_by:
    - chatcms-gw8g
    - chatcms-fdo0
---

## Purpose

One quiet indicator that tells the truth about saving and syncing, so nobody
has to think about git.

## Scope

- The five states from DESIGN.md section 5.10: saved, saving, offline,
  updates from others, and the same field edited by two people.
- Offline means saved on this device, and syncing on reconnect.
- Updates from others are pulled in automatically and announced with the dark
  toast, with a Show link.
- A same-field conflict renders inline on the field as a question, with
  Keep mine, Use theirs and Compare.
- Implementer-only technical details on hover or tap: branch, head, remote
  state, in mono.
- The Sync status screen, and the indicator in both the desktop top bar and the
  phone header.

## Done when

- Every state is reachable in a test, including offline and the same-field
  conflict.
- Saving never blocks typing.
- No git word reaches an editor, while the implementer tooltip is accurate.

## References

- `docs/BRIEFING-M01.md` sections 4.7 and 4.13.
- `docs/design/DESIGN.md` section 5.10.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-sync-status`.

This is the piece the git sync engine deliberately left undone. The engine
still answers in files; `internal/sitesync` turns a file into a page and a
page into questions about its fields, using the same model and form loader the
editor uses. Nothing about conflicts moved into the engine and nothing about
git moved out of it.

- All five states from the designs, each reachable in a test.
- Being unreachable is a standing, not a failure: the work stays here.
- A same-field difference is a question about that field, with keep mine, use
  theirs and a word-by-word compare.
- **Doing nothing keeps your own work.** A field nobody answers about keeps
  what was written here; taking the other version is the deliberate act.
  Anything else would mean a moment's inattention loses what somebody just
  typed.
- Everything has to be answered before anything is written, because a
  half-settled exchange is a state nobody could reason about afterwards.
- Technical detail is **left out** of the answer for somebody who may not see
  it, not merely unrendered: a screen that hides what it was sent is one
  view-source away from being wrong about who may see what.

One finding: the first run reported a PNG as a difference in "Page text" and
offered a screen of mojibake as the two answers. A file is now asked about
field by field only when it holds words; anything else is a choice between
whole versions.

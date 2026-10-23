---
# chatcms-8d4z
title: Sync status and conflict surfacing
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T15:18:26Z
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

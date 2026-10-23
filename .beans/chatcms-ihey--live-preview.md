---
# chatcms-ihey
title: Live preview
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T11:29:42Z
parent: chatcms-8vem
blocked_by:
    - chatcms-gw8g
---

## Purpose

Live preview of the site while editing, fast enough to feel immediate, and not
left running when nobody is looking.

## Scope

- `internal/preview`: a Hugo server per editing session, started on demand,
  with the pinned Hugo from `CHATCMS_HUGO_BIN`.
- Idle shutdown with a timeout, and a cap on concurrent preview servers.
- Reverse proxying the preview through the app so it is reachable under the
  app's own origin and respects the session and the policy function.
- The preview column from DESIGN.md section 5.6: address pill, desktop and
  phone device toggle, framed panel, and the Form / Side by side / Preview
  segmented control.
- The focused field highlighted in the preview.

## Done when

- A change in a form appears in the preview within a few seconds, which is
  acceptance test 5, asserted without sleeps.
- An idle preview is stopped and its resources released, proven by test.
- A viewer can preview and cannot edit.
- Preview works at 390 px through the phone layout.

## References

- `docs/BRIEFING-M01.md` section 4.8.
- `docs/design/DESIGN.md` section 5.6.

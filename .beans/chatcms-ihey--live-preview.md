---
# chatcms-ihey
title: Live preview
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T21:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-live-preview
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-live-preview`.

Acceptance test 5 passes: a change saved to a site shows in its preview within
seconds, asserted by asking until it appears rather than by sleeping.

- Hugo binds this machine only and is reached through the app's own address.
  A preview listening on its own would be a customer's unpublished work with
  no sign-in in front of it.
- Nothing tells the preview to rebuild: Hugo already watches the working copy
  that saving writes to.
- An idle preview is stopped, proven in a test that moves the sweeper's clock
  rather than waiting a quarter of an hour. One sweep loop, not a timer each,
  so there is one thing to stop when the app stops.
- A cap on how many run at once, and a preview whose working copy moved is
  stopped and started afresh rather than serving files that are gone.
- Every preview is stopped on both shutdown paths; stopping twice is safe.
- The preview column in the editor, with desktop and phone widths.

Field to preview linking is the one item from the form epic still outstanding.
It needs the preview to accept instructions from the editor and the site to
carry markers Hugo would render, which is more than this epic. Everything it
needs now exists.

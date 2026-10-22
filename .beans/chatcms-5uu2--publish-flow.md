---
# chatcms-5uu2
title: Publish flow
status: in-progress
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T20:12:52Z
parent: chatcms-8vem
blocked_by:
    - chatcms-2c0p
    - chatcms-gw8g
---

## Purpose

Publishing: the one deliberate action in the product. Show in plain language
what changed, then push.

## Scope

- `internal/publish`: compute what changed since the last publish, as pages and
  fields rather than as diffs, and record each publish.
- An inline word diff per change, expandable, with removals struck through and
  additions highlighted.
- An optional note, "What's in this update?", which History shows later.
- Push to main or to the configured branch, pushing either the Hugo source or
  the HTML built by the runner, per the project's publishing configuration.
- The four publish states from DESIGN.md section 5.8: review, in progress with
  its four plain steps, done, and error.
- The sticky phone publish bar and the phone publish screen.

## Done when

- Publishing shows what changed, then pushes, and the external pipeline
  deploys, which is acceptance test 7.
- "You can keep editing" is true: editing during a publish does not corrupt the
  publish or the edit.
- A failed publish leaves the live site untouched and says so.
- No git words appear anywhere in the editor-facing publish flow.

## References

- `docs/BRIEFING-M01.md` section 4.11.
- `docs/design/DESIGN.md` section 5.8.

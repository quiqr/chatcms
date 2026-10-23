---
# chatcms-0aeg
title: History and rollback
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T07:13:51Z
parent: chatcms-8vem
blocked_by:
    - chatcms-5uu2
---

## Purpose

Let a content creator see what was published when, and go back to an earlier
version, in plain language and with no fear.

## Scope

- `internal/history`: list versions from the publish records and git history,
  titled by the publish note or a generated summary, grouped by period, with
  the current one marked live.
- Show what a version contains, and preview the site as it was.
- Roll back with a confirmation dialog that says what will happen, reassures
  that nothing is lost, warns when unpublished changes will be set aside, and
  offers "Publish right away", default on.
- The History and rollback screens from DESIGN.md section 5.9.

## Done when

- Rolling back to an earlier version works and is itself reversible, which is
  acceptance test 8.
- Unpublished work is set aside rather than lost, proven by test.
- The timeline reads as plain language with no SHAs for editors.

## References

- `docs/BRIEFING-M01.md` section 4.12.
- `docs/design/DESIGN.md` section 5.9.

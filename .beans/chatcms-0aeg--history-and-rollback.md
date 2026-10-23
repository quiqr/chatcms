---
# chatcms-0aeg
title: History and rollback
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T10:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-history-and-rollback
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-history-and-rollback`.

Acceptance test 8 passes end to end: an integration test publishes twice,
goes back to the first version, reads the customer's remote to check the live
site, then goes back again to the version it moved away from and checks the
site is restored.

**Going back makes a version; it removes none.** The alternative, moving the
site's line backwards, destroys work and cannot be undone. Making a new
version holding the old content means the history keeps everything and going
back is reversible by going back again, with no special machinery for the
undo: it is the same operation.

**"Set aside" is accurate rather than soothing.** Unpublished work stops
being in the site when somebody goes back past it, and the confirmation says
how much will be. It is not deleted: the version holding it stays in the
history. A test writes a draft, goes back past it, checks it is gone from the
site, then goes back to the version that held it and reads the draft again.

**A title comes from the note somebody wrote, or the version's own message,
or a summary of what changed.** The third exists because some messages are
written by machinery, and "Merge branch 'main' into main" is both meaningless
to a content creator and full of words the copy rules forbid.

**No version identifier is ever shown.** It is carried so a request can name
it, and a test walks the timeline and fails if one appears in anything a
person reads.

**The git engine learned four things**: a version holding an earlier
version's content, what one version changed, the site at a version, and the
site just before one. The last two are read-only filesystems that the preview
epic will take from here.

Two findings worth recording:

- `fstest.TestFS` over the new filesystem found two real defects: listing
  entries reported no size, and a second `ReadDir(-1)` repeated itself rather
  than returning nothing. Both would have bitten the preview epic.
- My own identifier validation blocked an internal lookup of `HEAD`, so the
  check for "going back would change nothing" silently never fired. Input
  from outside and lookups this package writes itself are now separate.

One unrelated fix: a frontend test was flaking on the build machine because a
screen's wait has a one-second default and took longer there. The whole suite
now waits five seconds, which is still short enough that a genuinely stuck
test fails rather than hangs.

Coverage 84.4% overall, `internal/history` 85.5%, `internal/httpapi` back to
89.6% after direct handler tests for both new endpoint groups.

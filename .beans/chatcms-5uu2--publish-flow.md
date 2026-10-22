---
# chatcms-5uu2
title: Publish flow
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T00:30:00Z
openspec-link: openspec/changes/archive/2026-10-22-add-publish-flow
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-22-add-publish-flow`.

Acceptance test 7 passes end to end, both ways a project can publish: an
integration test edits a page, reads the review, presses publish, and then
reads the customer's remote to check what landed there.

**What changed, as pages and fields.** The review walks each differing file,
asks the model which page it is, parses the front matter on both sides and
compares field by field. A field is named as the model names it, with its
parent in front when nested, and a field with no title gets its key turned
into words. A file the model says nothing about is still listed, named from
its path: somebody about to publish should see everything that is going out.

**The word diff** compares words rather than characters, and the test that
holds it honest reads the comparison back both ways. Keeping what was kept
and removed must give the published text; keeping what was kept and added
must give the text as it is now.

**The git engine learned four things**, all in the interface's own
vocabulary: what differs from the published site, what the live site holds
for a file, publishing to a named line, and publishing built pages.

**"You can keep editing" is true, not hopeful.** Publishing built HTML the
obvious way would check out another branch and take the working copy away
from whoever is typing. `PublishPages` builds the tree with git's own
plumbing instead, in a temporary index, never reading or writing the working
copy. There is a test that saves a page while a publish is running and
checks that both survive.

**A failure carries two halves kept apart** all the way from the workflow
outcome through the record and the HTTP view to two places on the screen: a
sentence for whoever pressed the button, and the raw output. The screen test
asserts no git word reaches what an editor reads, using failures that return
real git noise so it would catch a leak rather than assume one cannot
happen. Who sees the raw half is `policy.Can` with `see-technical-detail`.

**One publish at a time**, refused by a unique index over a project's running
publishes rather than by a check two requests could both pass. Two publishes
racing to the same line is the one failure here that could leave a customer's
live site in a state nobody asked for.

**The screen** does all four states the designs draw, with the inline word
diff struck through and highlighted, the optional note, the step list, and
technical details behind a disclosure for implementers. The publish bar
sticks to the bottom on a phone.

One finding worth recording: a test assumed `Create` published, and it does
not. The sequence is create, then publish, and the test now walks it.

Coverage 84.2% overall, `internal/publish` 83.0%, `internal/gitsync` 82.8%.

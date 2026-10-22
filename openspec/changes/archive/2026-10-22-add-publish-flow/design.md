# Design

Epic: `.beans/chatcms-5uu2--publish-flow.md`

## What the git engine had to learn

Publishing asks the repository three questions it could not answer:

| Question | Method |
| --- | --- |
| What differs from the site people see? | `Changes` |
| What does the live site hold for this file? | `ReadPublished` |
| Send this to that line, as source or as pages | `PublishTo`, `PublishPages` |

All four speak in sites and lines, not in commits and branches, which is the
rule `docs/ARCHITECTURE.md` already sets for this interface.

## Built pages without touching the working copy

The epic asks that "you can keep editing" be true during a publish. For a
project that publishes built HTML that rules out the obvious implementation,
which is to check out another branch, empty it, copy the pages in and commit.
That takes the working copy away from whoever is typing.

`PublishPages` builds the tree with git's own plumbing instead: each page is
written to the object store, an index is assembled somewhere else entirely,
a tree is written from it, a version is made from that tree with the previous
publish as its parent, and that version is pushed. The working copy is never
read and never written. The index lives in a temporary directory named by an
environment setting, so there is no second index to collide with either.

That is also why a page no longer produced disappears from the line: the tree
is built from what was given, not patched onto what was there.

## What changed as pages and fields

A file-level answer would have been a quarter of the work and the wrong
product. The review walks each differing file, asks the model which page it
is, parses the front matter on both sides, and compares field by field.

Three decisions fell out of that:

**A file the model says nothing about is still listed.** A layout, a script,
an image outside the model. It is named from its path and badged as a change.
Somebody about to publish should see everything that is about to go out.

**A field is named as the model names it**, with its parent in front when it
is nested, so "Address, Street" reads as one thing. A field with no title
gets its key turned into words, because Quiqr models mix camelCase,
kebab-case and snake_case and an editor should see none of them.

**The comparison is by word, not by character.** A person wants to know which
words changed. The test that holds it honest reads the comparison back both
ways: keeping what was kept and removed must give the published text, and
keeping what was kept and added must give the text as it is now.

Very long text is marked replaced wholesale rather than compared. The
comparison is quadratic, and a page body is not worth a second of somebody's
time.

## One publish at a time, and the database is what says so

Two publishes racing to the same line is the one failure here that could
leave a customer's live site in a state nobody asked for. A unique index over
the running publishes of a project is what refuses the second, rather than a
check in Go that two requests can both pass.

## Why a publish is one request

The designs draw four steps with a progress bar, which suggests something
watched. Milestone 1 publishes a small site in a few seconds, so the request
carries the publish and the answer carries every step with its own outcome.
The record is written at the start and finished at the end, so moving this
onto the job queue later costs no schema change and no change to the screen's
shape.

## The two halves of a failure

A sentence for whoever pressed the button, and the raw output. They are
separate fields the whole way out: through the workflow outcome, through the
record, through the HTTP view, to two different places on the screen. The
screen test asserts that no git word reaches what an editor reads, with
failures that return real git noise so the test would catch a leak rather
than assume one cannot happen.

Who may see the raw half is `policy.Can` with `see-technical-detail`, asked
in the service, which empties the field before it leaves.

## What this change does not do

History and rollback is its own epic; this change records publishes and lists
them, which is what that epic will read. The progress bar during a publish
is a single answer rather than a stream, as above.

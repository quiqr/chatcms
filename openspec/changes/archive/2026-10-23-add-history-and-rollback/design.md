# Design

Epic: `.beans/chatcms-0aeg--history-and-rollback.md`

## Going back makes a version; it does not remove any

Two ways to put a site back to how it was:

| | What it does | Can it be undone |
| --- | --- | --- |
| Move the site's line backwards | Work after that point stops existing | No |
| Make a version holding the old content | Everything stays in the history | Yes, by going back again |

The second is the only one worth offering to somebody who is nervous about
pressing it. It is implemented by reading the old version's tree into the
index and the working copy and then making a version of that, so the result
is exactly the old content, including files the old version did not have
being gone.

Going back is therefore reversible without any special machinery: the version
somebody moved away from is still in the timeline, and going back to it is the
same operation again. The test does precisely that, and then checks the live
site.

## "Set aside" means set aside

Unpublished work stops being in the site when somebody goes back past it. The
word in the designs is "set aside", and it is accurate rather than soothing:
the version holding that work is still in the history, and going back to that
version brings the work out again. There is a test that writes a draft, goes
back past it, checks it is gone from the site, then goes back to the version
that held it and reads the draft again.

The confirmation says how many unpublished changes will be set aside, because
that is the one thing somebody would otherwise have to guess at.

## Where a title comes from

Three sources, in order:

1. The note somebody wrote when publishing. They wrote it to be read.
2. The version's own message, when a person wrote it.
3. A summary of what changed: "2 pages and 1 image changed".

The third exists because some messages are written by machinery, and "Merge
branch 'main' into main" is both meaningless and full of words the copy rules
forbid. A message is treated as machinery's when it begins with one of those
words.

## Attaching a note to a version

A publish record does not name a version, because publishing sends whatever is
waiting rather than one thing. The version a publish put live is the most
recent one made before that publish finished, which is what the server uses to
attach notes. It is a derivation rather than a stored link, so nothing has to
be migrated and a publish that failed attaches nothing.

## The site as it was

`At` gives a read-only filesystem over a version's tree: the listing comes
from git, each file's content is read only when it is opened. `Before` gives
the same for the state just before a version, which is the other side of the
comparison when showing what one version changed, and an empty site before the
first version rather than an error.

Writing a filesystem by hand is the kind of thing that passes a hand-written
test and fails a real caller, so `fstest.TestFS` runs against it. It found two
things: entries reported no size, and a second `ReadDir(-1)` repeated itself
instead of returning nothing. Both would have bitten the preview epic rather
than this one.

## Reusing the comparison

Showing what one version changed is the same question as showing what is
waiting to be published, asked of a different pair of trees. Rather than write
it twice, `internal/history` hands `internal/publish` an adapter whose three
questions are answered from one version instead of from the published site.
Two helpers there became exported for it.

## What this change does not do

Previewing the site as it was needs a preview server, which is its own epic.
This change gives that epic the site at a version as a filesystem, which is
what it will need.

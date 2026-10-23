# Design

Epic: `.beans/chatcms-8d4z--sync-status-and-conflict-surfacing.md`

## The piece the git engine left

When the git sync engine shipped it said plainly what it had not done:

> **Conflict resolution is file-level here.** The engine reports a conflicted
> file with its three versions and accepts replacement content. Turning that
> into the per-field question the design asks for needs the form layer, and
> belongs with the form epic.

This is that. The engine still answers in files; this package turns a file
into a page, and a page into questions about its fields, using the same model
and the same form loader the editor uses. Nothing about conflicts moved into
the engine, and nothing about git moved out of it.

## Only words get field questions

A picture, parsed as a document, is a document whose body is the binary. The
first run of this reported a PNG as a difference in "Page text" and offered a
screen of mojibake as the two answers. A file is now only asked about field by
field when the model names it as a page or its extension says it holds words;
anything else is a choice between whole versions.

## Doing nothing keeps your own work

A field nobody answers about keeps what was written here. Taking the other
version is the deliberate act, never the default. The alternative would mean
that a moment's inattention loses the thing somebody just typed, which is the
single worst outcome this screen could have.

Everything has to be answered before anything is written, because a
half-settled exchange is a state nobody could reason about afterwards.

## Technical detail is left out, not hidden

Where somebody may not see the branch, the remote and the head, those fields
are absent from the answer rather than present and unrendered. A screen that
hides what it was sent is one view-source away from being wrong about who may
see what.

That block is also the one place in the whole product where these words are
allowed, which is why it carries the implementer marker the interface check
knows about.

## Offline is reported, not declared

It is what the engine says when the remote cannot be reached. The app does not
keep its own idea of being online, so the two cannot disagree.

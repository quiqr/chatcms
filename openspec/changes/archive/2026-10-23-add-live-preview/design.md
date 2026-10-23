# Design

Epic: `.beans/chatcms-ihey--live-preview.md`

## Hugo binds this machine only

A Hugo server listening on its own would be a customer's unpublished work with
no sign-in in front of it. It binds `127.0.0.1` on a port nobody is told, and
the app forwards to it after asking who somebody is. The address a browser
uses is the app's own, under `/preview/<organisation>/<project>/`.

That is also why a preview is not simply linked to: the thing that knows who
somebody is has to be in the path.

## Changes show because Hugo is already watching

Nothing in ChatCMS tells a preview to rebuild. Hugo watches the working copy,
which is the same directory saving writes to, so a save appears by itself.

The test waits for it to show rather than sleeping for a guess: it asks until
the new words appear or twenty seconds pass. A sleep long enough to be
reliable would make the suite slow, and one short enough to be quick would
make it flaky.

## Idle, and how it is tested

An idle Hugo server holds memory for nothing. Previews are swept by one loop
rather than a timer each, so there is one thing to stop when the app stops and
no timer can outlive what it was counting for.

Testing "nobody has looked at this for a quarter of an hour" would otherwise
take a quarter of an hour. The sweeper asks a function for the time, and a
test replaces it, so the whole thing runs in milliseconds and the real
behaviour is still what is being tested.

## A moved working copy

A preview serving a directory that was made afresh is serving files that are
no longer there. Asking for a preview whose site has moved stops the old one
and starts another. Serving a customer stale pages is worse than a moment's
wait.

## What this change does not do

Field to preview linking, where the field being edited is outlined in the
preview, needs the preview to accept instructions from the editor and the site
to carry markers Hugo would have to render. That is more than this epic, and
it is the one item from the form epic still outstanding. Everything it needs
now exists.

# Design

Epic: `.beans/chatcms-ac9e--media-library.md`

## A file's name is not evidence

What a picture is comes from reading its first bytes, never from its
extension. The name arrives from whoever sent it and a file called `.png`
that is a text document would be published as a broken picture on a
customer's site. The name is then rewritten from what the bytes turned out to
be, so a photograph misnamed by a phone still ends up named for what it is.

SVG is the exception that has to be handled rather than sniffed: it is text,
so the usual detection reports it as plain text or XML. It is recognised by
its own opening tag instead.

## Nothing is replaced

A picture arriving with the name of one already there is put beside it as
`name-2`. Replacing would silently change every page using the first, which
is the kind of thing somebody discovers a week later on the live site.

## Descriptions live in the site

A picture's bytes have nowhere to keep what it shows, and a database would
mean a site's descriptions did not travel with the site. They go in
`quiqr/media.yaml`, written with names in order so that the same descriptions
always make the same file: otherwise every save would show as a change in the
publish review.

## Usage is a search, not an index

Which pages use a picture is worked out by reading the content each time. An
index would be a second source of truth and would be wrong the first time
anybody edited a page outside ChatCMS, which is a thing this product
explicitly supports. Every shape a Hugo site writes an address in is looked
for, not only the shape ChatCMS would have written.

## A small copy is a cache in the strict sense

Named after a hash of the bytes it was made from, so:

- the same picture is made small once;
- a changed picture gets a new name by itself, with no invalidation to get wrong;
- the whole directory can be deleted at any moment and is rebuilt from the
  site in git alone, which is what "rebuildable from git" has to mean.

Nothing in it is backed up and nothing migrates. It is written beside and
moved into place, so two requests for the same new picture cannot leave half
a file behind.

Resizing needs an algorithm the standard library does not have, which is why
`golang.org/x/image` is a dependency. A kind that cannot be decoded is given
back whole rather than reported as broken: it is still a picture the site
holds.

## The picker

An image field opens the library and fills itself in with what is chosen. The
library does not know about fields: the editor passes down what to do with a
chosen picture, so the same screen serves both browsing and picking.

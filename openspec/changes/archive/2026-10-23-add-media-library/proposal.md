## Why

A site is pictures as well as words, and an image field in a form is useless
without somewhere to pick from. Epic `chatcms-ac9e`, and the image half of
acceptance test 4.

## What Changes

- **A picture library**: every picture in the site, searchable, with a filter
  for the ones nothing uses, dropped or chosen from the machine.
- **What a picture is comes from its own bytes**, not from what it was called,
  so a file named `.png` that is not one is refused rather than published as a
  broken picture.
- **A picture already there is not replaced**: a second one of the same name is
  put beside it, because replacing would change every page using the first.
- **A description for screen readers**, kept in the site so it travels with it.
- **Which pages use a picture**, found by looking, so it cannot go stale.
- **Removing warns when something still uses it**, and says how many pages
  would have a gap.
- **Small copies for the grid**, made once and kept, named after the bytes
  they came from so the cache can be thrown away at any time.
- **An image field in a form opens the library** and fills itself in.

## Capabilities

### New Capabilities

- `media/library`: what a site's pictures are, what is known about each, and
  what a person can do with them.
- `media/pictures`: what ChatCMS will take as a picture, what it names it, and
  what it gives back.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/media`, `web/src/routes/MediaPage.tsx`, the picture
  field in `web/src/fields/`.
- **Dependency added**: `golang.org/x/image`, for making a small copy well.
  The standard library decodes pictures but cannot resize one.
- **Schema**: none. Pictures and their descriptions live in the site.

## Assumptions

- **A picture is what its bytes say, never what its name says.** A name is
  taken from the sender and cannot be trusted; a misnamed file would be
  published as a broken picture on a customer's site. The name is then
  rewritten from what the bytes turned out to be.

- **Descriptions live in the site, in `quiqr/media.yaml`.** A picture's own
  bytes have nowhere to keep one, and a database would mean a site's
  descriptions did not travel with the site. The file is written with its
  names in order, so the same descriptions always make the same file and a
  publish shows no spurious change.

- **What uses a picture is found by looking, every time.** An index would be a
  second source of truth and would be wrong the first time anybody edited a
  page outside ChatCMS. Looking costs a read of the content, which for a site
  of this size is nothing.

- **A small copy is a cache and nothing more.** It is named after a hash of
  the bytes it was made from, so the directory can be deleted at any moment
  and is rebuilt from the site in git alone. Nothing in it is backed up and
  nothing migrates.

- **The largest picture a site will take is 25 MB.** A photograph from a phone
  is a few megabytes. Much more than that in git is a mistake rather than a
  choice, and the refusal says the size in words a person can act on.

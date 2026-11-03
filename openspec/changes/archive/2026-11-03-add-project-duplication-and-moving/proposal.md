## Why

The organisation area draws four acts on a whole site and offers none of
them: duplicate, move to another organisation, download as zip, and archive.
They are drawn disabled and say they are not built, which was the right
answer while the area was being built and is the wrong one now.

Epic `chatcms-hq8m`.

Bean: `.beans/chatcms-hq8m--a-project-duplicated-moved-zipped-or-archived.md`.

## What Changes

- **Duplicate**: a separate copy in the same organisation, with a name of its
  own. Pages, media and the content model always; the people and the
  publishing settings only if asked for.
- **Move**: a site goes to another organisation. Its website, its publishing
  and its history go along unchanged; the people on it keep their roles and
  become members of the destination.
- **Download as zip**: the site's files, as they are.
- **Archive**: a site stops counting against the organisation's limit and
  stops being in the way, without anything being deleted.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/projects`: four acts on a whole site, and who may do each.

## Impact

- **New**: four endpoints, the dialogs the design draws, and a copy of a
  working copy that shares nothing with its original.
- **Changed**: the menu in the organisation area offers them.

## Assumptions

- **A copy shares nothing with its original, including where it publishes.**
  The copy is made as a new working copy from the original's files, with no
  remote unless somebody deliberately ticked the publishing settings. A copy
  that published where the original publishes would overwrite a live website
  the first time anybody pressed Publish, which is why the design has that
  box off by default and why this does the same thing underneath.

- **The copy must be given its own name.** The dialog opens with the
  original's name already in it and already refused, which is the design's
  way of saying "think about this" rather than making somebody delete a word
  before they can type.

- **Moving touches the database and nothing else.** A site's working copy and
  its remote are named by the project, not by the organisation, so the files
  do not move and git is not told anything. What changes is which
  organisation holds it and who is a member of that organisation.

- **The people come with the site.** Somebody who edits a site and is not a
  member of the organisation it moves to would lose it; so they are made
  members of the destination, counted against its limit before the move runs.

- **The zip is the site's files as they are.** The design says pages, media
  and the content model, which is what a site is made of; rather than pick
  three folders and be wrong about the fourth, it is the working copy
  without its git directory.

- **Archiving is not a lesser deleting.** Nothing is removed, nothing stops
  being readable, and what is published stays published. It is a way to say
  "not now" that the project limit respects.

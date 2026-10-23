## Why

A content creator who cannot undo is a content creator who edits carefully and
slowly. Seeing what was published when, and being able to go back without
fear, is what makes the rest of the product safe to use. Epic `chatcms-0aeg`,
acceptance test 8.

## What Changes

- **A timeline** of what was published when, grouped into stretches of time a
  person would name, titled by the note somebody wrote or by a summary of
  what changed, with the version the live site holds marked.
- **What a version contains**, as pages and fields, the same shape the review
  before publishing uses.
- **The site as it was**, readable without disturbing what anybody is editing,
  which is what previewing an old version needs.
- **Going back** makes a new version holding an earlier version's content. It
  throws nothing away, so it can itself be undone.
- **A confirmation that says what will happen**: which version, when it was,
  how much unpublished work will be set aside, and that nothing is lost, with
  "Publish right away" on unless it is turned off.
- **The History screen**, working at 390 px.

## Capabilities

### New Capabilities

- `history/timeline`: what is listed, in what order, and in what words.
- `history/going-back`: what going back does, what it says first, and what it
  leaves behind.

### Modified Capabilities

- `sync/repository`: the git engine learns to make a version holding an
  earlier version's content, to say what one version changed, and to read the
  site as it was at a version and just before one.

## Impact

- **New code**: `internal/history`, `web/src/routes/HistoryPage.tsx`.
- **Changed**: `internal/gitsync` gains four methods; `internal/auth` gains a
  person lookup by identifier; two helpers in `internal/publish` become
  exported so the history can name files the same way.
- **Schema**: none. The history is the site's own versions plus the publish
  records that already exist.
- **Depends on**: `add-publish-flow`.

## Assumptions

- **Going back makes a version rather than removing them.** The alternative,
  moving the site's line backwards, destroys work and cannot be undone. Making
  a new version with older content means the history keeps everything, going
  back is reversible by going back again, and somebody who went back by
  mistake loses nothing.

- **"Set aside" is honest, not a euphemism.** Unpublished work stops being in
  the site when somebody goes back past it. It is not deleted: the version
  holding it is still in the history, and going back to that version brings it
  out again. The confirmation says how much will be set aside, and a test
  proves the work can be brought back.

- **A publish note is attached to the version that was live when it
  finished.** A publish record does not name a version, because publishing
  sends whatever is waiting. The most recent version made before the publish
  finished is the one it put live, which is the version the note is about.

- **The version identifier is carried but never shown.** It is what a request
  names. A test walks the timeline and fails if an identifier appears in
  anything a person reads.

- **Previewing the site as it was is left to the preview epic.** This change
  gives it the site at a version as a read-only filesystem, which is what a
  preview server needs. Starting one is that epic's work.

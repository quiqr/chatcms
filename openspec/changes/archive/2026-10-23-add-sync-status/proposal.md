## Why

A content creator who cannot tell whether their work is safe will not trust
the product. And when two people change the same thing, somebody has to
choose, in words about their page rather than about files. Epic
`chatcms-8d4z`, and the piece deliberately deferred when the git sync engine
shipped.

## What Changes

- **The five states from the designs**: saved, saving, offline, somebody else
  updated this site, and two people changed the same thing.
- **Being unreachable is a state, not a failure.** Work stays here and goes
  out when the site can be reached again.
- **What arrived is announced with whose it was.**
- **A same-field difference is a question about that field**: keep mine, use
  theirs, or compare, with the words that differ marked.
- **A file that cannot be shown field by field** is a choice between one whole
  version and the other, said plainly.
- **Technical detail for implementers only**, in one place, which is the one
  place in the product where these words belong.

## Capabilities

### New Capabilities

- `sync/standing`: how a site stands and in what words.
- `sync/settling`: what happens when two people changed the same thing.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/sitesync`, `web/src/routes/SyncPage.tsx`.
- **Changed**: two helpers in `internal/publish` become exported, so a
  conflict is named the same way the publish review names a change.
- **Schema**: none.

## Assumptions

- **Only a file that holds words is asked about field by field.** A picture
  parses as a document whose body is the binary, and showing that would be a
  screen of mojibake. Anything that is not content the model names is a choice
  between whole versions instead.

- **A field nobody answers about keeps what was written here.** Doing nothing
  should never throw away the work of the person doing it. Taking the other
  version is the deliberate act.

- **Every page has to be answered before anything is written.** A half-settled
  exchange is a state nobody could reason about afterwards.

- **Technical detail is never sent to somebody who may not see it.** It is
  left out of the answer entirely rather than hidden by the interface, so that
  a view of the network traffic shows the same thing the screen does.

- **Offline is detected, not declared.** It is what the engine reports when
  the remote cannot be reached, so it cannot disagree with reality.

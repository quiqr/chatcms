## Why

Everything a platform admin can do is stacked on one screen: starting an
organisation, the list of them, who administers the installation, somebody
who has left, every person, making one, and the record of what was done. It
grew that way an epic at a time, and it now reads as a drawer rather than a
screen.

The hand-off of 2 November gives the platform admin an area: a dark shell so
it is obvious where you are, a list of organisations with what each one holds
and how full it is, and one organisation opened on a screen that holds
everything about it.

Epic `chatcms-tr45`.

Bean: `.beans/chatcms-tr45--the-platform-admin-area-redrawn-a-dark-shell-organ.md`.

## What Changes

- **A dark shell at `/administration`**, with a panel saying you are in it,
  Organisations, Users and History with counts, and the way back.
- **Organisations**: search, filter chips that the server answers, and a
  table with each organisation's admins or a warning that it has none, its
  usage against its limits, when it was made and by whom, and the way in.
- **One organisation**: its details, its limits with a stepper, its sites,
  its members, and removing it.
- **Removing an organisation takes its sites with it**, behind a dialog that
  lists what goes and asks for the organisation's name to be typed.
- **Who made an organisation is kept**, so the screen can say it.
- The old stacked screen goes. The people half of it moves to `Users` under
  the new shell, unchanged, for `chatcms-503w` to redraw.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/administration`: the area, what it lists, and what removing an
  organisation now does.

## Impact

- **New**: a shell, two screens, a column for who made an organisation, and
  searching and filtering answered by the server.
- **Changed**: removing an organisation no longer refuses while it holds
  sites; `AdministrationPage` is retired.

## Assumptions

- **Removing an organisation now removes its sites with it.** The epic that
  built that screen refused while any site remained, so that letting go of
  several customers' work could not be one confirmation. The hand-off draws
  the opposite deliberately: the dialog lists what goes, says live websites
  and what is in git are untouched, and asks for the organisation's name to
  be typed. That is the protection the refusal stood in for, and it is the
  design's call rather than mine. What is in git is still never touched.

- **The filter chips are the server's answer.** "At a limit" and "No admin
  yet" are questions about every organisation, and a thousand of them must
  not be sent to a browser to be counted there.

- **Who made an organisation is a new column**, null for the ones made
  before it. The record already says who did it, but a record is a list of
  acts, and the screen wants a fact about a thing.

- **History stays, though the hand-off does not draw it.** The product
  already writes down every administration act; dropping the screen that
  reads it would be a regression nobody asked for. It is the third item in
  the nav rather than a fourth card on a screen.

- **Users keeps what it has until `chatcms-503w`.** Moving the people list
  under the new shell unchanged is what retires the stacked screen now,
  rather than leaving the old one beside the new area while the next epic
  runs.

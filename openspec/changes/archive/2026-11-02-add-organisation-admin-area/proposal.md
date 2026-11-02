## Why

An organisation admin has no screens of their own. They manage the people of
an organisation through a tab on a project's People and roles screen, which
means the organisation is only reachable through one of the things inside it,
and they have nowhere at all to see the organisation as a whole: what it
holds, how full it is, and which sites are in it.

The hand-off of 2 November gives them an area: a shell of its own, a members
screen with seats, and a projects screen.

Epic `chatcms-ufd3`.

Bean: `.beans/chatcms-ufd3--the-organisation-admin-area-its-own-shell-members.md`.

## What Changes

- **An area at `/o/{organisation}/manage`**, with a shell of its own: the
  organisation tile, Projects and Members with counts, a usage box saying how
  full it is and who sets that, and a way back to your own projects.
- **Members**: a seat pill, an invite card that takes an address and a role
  in the organisation, and a table of who is in it with the projects each one
  reaches, what they last did, and taking somebody out. The full state says
  so and stops offering the form.
- **Projects**: the organisation's sites as a list rather than a table, each
  with what it is, how many people it has, when it was last published, a
  status, a way in, and the menu the design draws.
- **The way in**: the organisation switcher gains "Manage <organisation>" for
  whoever administers one.
- **The people tab on a project keeps the project**, and the organisation
  half of that screen moves here, because an organisation is not a thing
  inside one of its projects.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/organisations`: an organisation has screens of its own, reachable
  by whoever administers it.

## Impact

- **New**: a shell, two screens, and an answer that says what one
  organisation holds in one request.
- **Changed**: the People and roles screen loses its Organisation tab; the
  switcher gains a way into the area.

## Assumptions

- **The menu items that belong to another epic are drawn and disabled.**
  Duplicate, Move, Download as zip and Archive are `chatcms-hq8m`. Drawing
  them now with nothing behind them would be a lie; leaving them out would
  mean rebuilding the menu twice. They are drawn, disabled, and say what they
  are waiting for, which is what the design's own "Soon" badge already does
  elsewhere in this product.

- **Removing somebody from the organisation takes their project roles with
  them.** Leaving a role on a project in an organisation somebody is no
  longer in would leave a reach nobody can see on any screen.

- **The organisation's own screens show the same usage as the platform
  admin's**, because both read `UsageFor`. The usage box says who sets the
  limits, so that somebody who cannot change them knows who can.

---
# chatcms-tr45
title: 'The platform admin area, redrawn: a dark shell, organisations and one organisation'
status: completed
type: epic
priority: high
created_at: 2026-11-02T20:05:43Z
updated_at: 2026-11-03T01:30:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-03-add-platform-admin-area
blocked_by:
    - chatcms-civ6
    - chatcms-qllv
---

## Purpose

The administration screen we shipped is one page with every act stacked on it.
The design replaces it with an area of its own: a dark shell that makes it
obvious where you are, a list of organisations with their usage, and one
organisation opened on a screen that holds everything about it.

## Scope

- **The platform admin shell** (section 3): a dark sidebar (`bg-secondary`),
  a "You're in Platform admin" panel with a shield, nav holding Organisations
  and Users with counts, "Back to my projects", and the person with the role
  "Platform admin". The content area stays the normal light page. The
  organisation switcher in the ordinary shell gains a "Platform admin" pill
  for whoever holds it, and loses "New organisation", because only a platform
  admin makes one and they do it from here.
- **Organisations** (5.14): search, New organisation, filter chips (All · n,
  At a limit, No admin yet), and a table with the organisation (tile, name,
  address, opening the detail), its organisation admins or "No admin yet ·
  invite pending" in the warning colour, members and projects usage cells,
  when it was created, and Manage.
- **One organisation** (5.15): a header with its tile, name and Save changes,
  then cards for Details (name, address `chatcms.eu/<slug>`, created by and
  when), Limits (a checkbox per limit with a stepper, the usage line and its
  bar, and the note that limits are optional), Projects (rows with Duplicate,
  Move and Zip, a platform admin being able to move to any organisation),
  Members (add by email with a role, and a table with role, projects and
  Remove), and Delete organisation as a danger card.
- **Deleting** (`admin-organisation-delete`): a dialog that lists what goes
  and asks for the organisation's name to be typed.
- At 1440 px and at 390 px.

## Out of scope

Users, which is its own epic because it is account work rather than
organisation work.

## Done when

- A platform admin does everything the old screen did, and the old screen is
  gone rather than left beside it.
- The filter chips answer from the server, so "At a limit" is right on a
  thousand organisations.
- Removing an organisation still refuses while it holds projects, and the
  dialog asks for its name first.
- The acceptance suite walks the area: listing, filtering, opening one,
  changing a limit, and deleting one.

## References

- `docs/design/reference/html/admin-organisations.html`,
  `admin-organisation.html`, `admin-organisation-delete.html`.
- What exists now: `web/src/routes/AdministrationPage.tsx`,
  `AdministeredOrganisationPage.tsx`, `internal/admin/`.

## Summary of Changes

The platform admin has an area rather than a page. The stacked screen that
grew an epic at a time is gone.

- **A dark shell** that says you are in it, with Organisations, Users and
  History beside their counts, and the way back. It has tokens of its own
  that sit outside the light and dark blocks, because an area whose point is
  that it looks different must not follow the theme into looking the same.
- **Organisations**: search and filter chips answered by the server ("at a
  limit", "no admin yet"), and a table with who administers each one, what it
  holds against its limits, and when it was made and by whom.
- **One organisation**: its details, its limits with a checkbox and a stepper
  (a limit is a number or nothing, never zero), its sites, its members, and
  removing it.
- **Removing an organisation takes its sites and memberships with it.** This
  reverses what `chatcms-d4f3` decided. The hand-off draws the dialog
  instead: it lists what goes, says live websites and what is in git keep
  working, and will not act until the organisation's name is typed. Nothing
  in git is touched either way.
- **Who made an organisation** is a column now, null for the ones made before
  it, because a record is a list of acts and the screen wants a fact about a
  thing.
- **The counts come with the answer that says you may be here**, so moving
  between the screens of the area is one request rather than three.
- Proved by five handler tests, three integration tests over listing,
  filtering and removal, and
  `test/e2e/24-platform-admin-area.spec.mjs` at 1440 px and at 390 px.

### What the gate caught

A 500 on the seeded database: the subquery for who made an organisation was
typed non-null by sqlc, and every organisation made before the column existed
has none. My own test had passed because it only made organisations the way
that fills the column in. It is a COALESCE now, and the test makes one with no
maker on purpose.

Then ten acceptance failures from the move itself: starting an organisation
is behind a button, the record is its own screen, and a search box was 42 px
tall on a phone because it sat in a bordered form rather than a label. The
fixtures gained `startsAnOrganisation` so the first of those is written once.

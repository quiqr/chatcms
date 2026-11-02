---
# chatcms-tr45
title: 'The platform admin area, redrawn: a dark shell, organisations and one organisation'
status: todo
type: epic
priority: high
created_at: 2026-11-02T20:05:43Z
updated_at: 2026-11-02T20:05:43Z
parent: chatcms-7r0u
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

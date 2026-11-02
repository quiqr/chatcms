---
# chatcms-ufd3
title: 'The organisation admin area: its own shell, members and projects'
status: completed
type: epic
priority: high
created_at: 2026-11-02T20:05:14Z
updated_at: 2026-11-03T00:30:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-02-add-organisation-admin-area
blocked_by:
    - chatcms-civ6
    - chatcms-qllv
---

## Purpose

An organisation admin has no screens of their own. Today they manage people
through a project's People and roles screen and have nowhere to see the
organisation as a whole. The design gives them an area with a shell of its
own, a members screen and a projects screen.

## Scope

- **The organisation admin shell** (`docs/design/DESIGN.md` section 3): a
  light sidebar like the project shell, with the organisation tile, a nav
  group "Manage organisation" holding Projects and Members with counts, a
  **usage box** showing members and projects against their limits with
  "Limits are set by the platform admin", "Back to my projects", and the
  person with the role "Organisation admin" beneath.
- **Members** (5.12): a seat pill in the top bar ("7 of 10 seats used", red
  when full), an invite card that takes an email, a role in the organisation
  (Member or Organisation admin) and project access (a project and a role),
  and a table with person, organisation role, their projects as chips
  ("Korenaar Bakery · Editor", "All projects" for an admin), last active, and
  Remove or Resend plus Cancel invite for one still waiting.
- **The full state** (`org-members-full`): a warning banner, "Request more
  seats", the invite fields muted and the button disabled.
- **Projects** (5.13): a list rather than a table so it reflows on a phone,
  each with its tile, "domain · n people · published when", a status badge,
  Open, and a menu: People and roles, Project settings, Rename, Duplicate,
  Move to another organisation, Download as zip, Archive, Delete project. A
  collapsed "Archived · n · doesn't count toward the limit" below.
- The limit pill in the top bar ("4 of 5 · 1 left") and New project beside it.
- At 1440 px and at 390 px.

## Out of scope

Duplicating, moving, zipping and archiving a project are their own epic. This
one draws the menu; that one makes the items work.

## Done when

- An organisation admin signs in, reaches their organisation's area from the
  switcher ("Manage <organisation>"), and manages members and projects
  without touching a project screen.
- An ordinary member cannot reach the area, refused by `policy.Can` rather
  than by a hidden link.
- The usage box and the pills read the same numbers as the platform admin
  screens.
- The acceptance suite covers inviting, changing a role, removing somebody,
  and the full state, at both widths.

## References

- `docs/design/reference/html/org-members.html`, `org-members-full.html`,
  `org-projects.html`.

## Summary of Changes

An organisation is a thing in itself now, with screens of its own at
`/o/{organisation}/manage`, rather than something reachable only through one
of the projects inside it.

- **One answer**, `GET /api/organisations/{organisation}/manage`, carrying
  the organisation, how full it is, everybody in it with the sites each one
  reaches, and its sites. One request because the shell needs the counts
  whichever screen of the area is open, and one query for the people because
  twenty members must not cost twenty-one.
- **A shell of its own**: the organisation tile, Projects and Members with
  counts, the usage box saying how full it is and that the platform admin
  sets that, and a way back to your own projects.
- **Members**: inviting with a role in the organisation, the sites each
  person reaches as chips, taking somebody out, and the full state saying so
  and withdrawing the form.
- **Projects**: the sites as a list that reflows, with the menu the design
  draws. Duplicate, Move, Download as zip and Archive are drawn disabled and
  say "Not built yet"; they are `chatcms-hq8m`.
- **The switcher** offers "Manage <organisation>" for each organisation
  somebody administers, decided by `policy.Can` rather than by the role, so
  that a platform admin who administers nothing is not offered one.
- **Taking somebody out of an organisation takes their roles on its sites
  with them.** A test proves it leaves their roles in other organisations
  alone.
- **The project People screen keeps the project.** The organisation half of
  it moved here, and `test/e2e/15-organisation-people.spec.mjs` moved with
  it rather than being deleted: the invitation still has to be sent and
  withdrawable.
- Proved by two handler tests, two integration tests over who may reach the
  area and what removal takes with it, and
  `test/e2e/23-organisation-area.spec.mjs` at 1440 px and at 390 px.

### Found on the way

The switcher now holds two entries per organisation, so a test matching
`/TechNative/` matched both. Two existing specs now match exactly or anchor
at the start, which is the honest fix rather than loosening the assertion.

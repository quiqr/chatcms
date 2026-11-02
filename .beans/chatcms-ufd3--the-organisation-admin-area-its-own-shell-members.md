---
# chatcms-ufd3
title: 'The organisation admin area: its own shell, members and projects'
status: in-progress
type: epic
priority: high
created_at: 2026-11-02T20:05:14Z
updated_at: 2026-11-02T22:14:58Z
parent: chatcms-7r0u
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

---
# chatcms-civ6
title: 'Three levels of role: platform admin, organisation admin, project roles'
status: in-progress
type: epic
priority: critical
created_at: 2026-11-02T20:04:48Z
updated_at: 2026-11-02T20:44:05Z
parent: chatcms-7r0u
---

## Purpose

The design of 2 November replaces the role model milestone 01 shipped. Three
levels, each managing the one below: platform admin over the installation,
organisation admin over one organisation, and project roles under that. The
project role "owner" disappears; what it could do belongs to the organisation
admin now.

Everything else in this milestone rests on this, so it goes first.

## Scope

- **Words.** "Administrator" becomes **platform admin** everywhere a person
  can read it. "Organisation owner" becomes **organisation admin**. The
  catalogue is one file, so this is a rename in `web/src/strings/` plus the
  spec and design text.
- **The project role "owner" goes.** `project_members.role` keeps
  implementer, editor and viewer. A migration moves every existing project
  owner to implementer, and the CHECK constraint drops the word.
- **An organisation admin reaches every project in their organisation**,
  which `policy.EffectiveRole` already does by raising an organisation owner
  to project owner. It raises to implementer now, and the actions that used
  to need project owner (manage billing, delete project, create project) are
  answered from the organisation role instead.
- **The permission table in `docs/design/DESIGN.md` section 6** is what
  `internal/policy` answers, including the two new rows: moving a project
  (platform admin anywhere, organisation admin only between organisations
  they also administer) and setting limits (platform admin only).
- **Roles belong to a place.** Somebody can be organisation admin in one
  organisation and editor on a project in another; nothing in the policy may
  read a person's role as a property of the person.
- The project People and roles screen offers Editor, Viewer and Implementer,
  and shows a fixed "Organisation admin" badge for somebody who holds that
  instead of a role select (`docs/design/DESIGN.md` 5.4).
- `docs/design/reference/html/hierarchy.html` is the picture of all of this.
  It is a reference board rather than a screen; the architecture document
  should say the same thing in words.

## Done when

- `policy.Can` answers exactly the table in section 6, proven by the test that
  walks every role against every action.
- No screen, string or spec says "owner" about a project or an organisation.
- An organisation admin opens a project nobody gave them a role on, and an
  ordinary member does not, proven by an integration test.
- The migration runs on a database holding project owners and leaves nobody
  without the access they had.

## References

- `internal/policy/policy.go`, `EffectiveRole` and the `allowed` table.
- `openspec/specs/tenancy/permissions/`, which states the old model.
- `db/migrations/00004_tenancy.sql`, the role constraints.

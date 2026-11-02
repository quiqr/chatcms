---
# chatcms-civ6
title: 'Three levels of role: platform admin, organisation admin, project roles'
status: completed
type: epic
priority: critical
created_at: 2026-11-02T20:04:48Z
updated_at: 2026-11-02T22:40:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-02-add-three-levels-of-role
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

## Summary of Changes

Three levels of role, each answered from a different place, which is what
keeps them from leaking into one another.

- **Migration `00013`**: every project owner became an implementer, every
  organisation owner became an admin, and both CHECK constraints dropped the
  old word. The order is drop the constraint, update the rows, add the new
  constraint; doing it any other way fails, and an integration test seeds a
  database at schema 12 with an owner of each kind to prove the whole thing.
- **`internal/policy`** holds three project roles. A new `overTheOrganisation`
  set carries billing, deleting a project, creating one and moving one; those
  are answered from the role held in that organisation, and no project role
  carries any of them. `EffectiveRole` raises an organisation admin to
  implementer, because there is no owner of a project to raise to.
- **The two actions the design names and nothing uses yet**, `move-project`
  and `set-limits`, are answered by level: moving by whoever administers the
  organisation, limits by the platform admin alone.
- **`Allowed` still reports what the organisation carries**, because the
  project settings screen asks it whether to offer Delete. A test pins the
  difference between what an implementer is offered and what an organisation
  admin is.
- **The words**, in English and Dutch: platform admin, organisation admin,
  member, implementer, editor, viewer. `role.owner` is gone from the
  catalogue, the project people screen offers the three roles it has, and
  what an organisation admin holds is described beside the table rather than
  offered in it.
- **`docs/ARCHITECTURE.md`** states the three levels and where each is
  answered from, which is what the hierarchy board draws.
- Proved by the whole role table, five new policy tests, two integration
  tests over what an organisation admin and a member reach, the migration
  test, and an acceptance test over the roles a screen offers.

### What the gate caught that nothing else did

Four acceptance specs still typed into "Who owns it" after the label became
"Who administers it": eight failures, one string. Every unit test passed
throughout. A rename is only finished when something reads the screen.

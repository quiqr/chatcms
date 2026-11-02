## Why

The design hand-off of 2 November replaces the role model milestone 01
shipped. Three levels, each managing the one below: a platform admin over the
installation, an organisation admin over one organisation, and project roles
under that. The project role "owner" goes; what it could do belongs to the
organisation admin now.

Three of the five other epics in this milestone rest on it, so it goes first.

Epic `chatcms-civ6`.

Bean: `.beans/chatcms-civ6--three-levels-of-role-platform-admin-organisation-a.md`.

## What Changes

- **Three project roles**: implementer, editor and viewer. A migration moves
  every project owner to implementer and the database stops accepting the
  word.
- **An organisation admin** is what an organisation owner was called, stored
  as `admin`, and reaches every project in their organisation as its
  implementer.
- **Billing, deleting a project and creating one** are answered from the
  organisation role rather than from a project role, because that is where
  the design puts them.
- **Two actions the table names and nothing does yet**: moving a project to
  another organisation, and setting an organisation's limits. The policy
  answers both; the acts themselves are `chatcms-hq8m` and `chatcms-qllv`.
- **The words a person reads**: platform admin, organisation admin, member.
  "Administrator" and "owner" are gone from every screen and every string.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/permissions`: the three levels, the table behind them, and what an
  organisation admin reaches.

## Impact

- **New**: a migration moving the stored roles.
- **Changed**: `internal/policy` and the table it answers from; every screen
  that names a role; the specs and the architecture document.

## Assumptions

- **A project owner becomes an implementer, and nothing more.** An
  implementer sets up the repository, publishing and the technical detail,
  which is what a project owner did day to day. What they lose is billing and
  deleting the project, which the design moves up a level deliberately. The
  alternative, making every project owner an organisation admin, would hand
  somebody who was given one project the keys to every project beside it.
  Whoever already owned an organisation keeps everything, because that role
  is renamed rather than changed.

- **The stored word changes with the spoken one.** An organisation role is
  stored as `admin` rather than `owner`. Leaving the database saying owner
  while every screen says admin is the kind of difference that is cheap today
  and expensive the first time somebody reads a query and believes it.

- **An organisation admin is an implementer on every project, not an owner
  of one.** There is no owner to be. Deleting a project is theirs because
  they administer the organisation, not because of the role they hold on the
  project, and `policy.Can` answers it that way.

- **The two new actions are added now, unused.** The table in section 6 is
  what this change delivers, and a table with two rows missing is a table
  that has to be read twice. Each is refused for everybody but the level the
  design names, and the epics that use them add no permission of their own.

- **The hierarchy board is documentation.**
  `docs/design/reference/html/hierarchy.html` is a reference picture, so what
  it says goes into `docs/ARCHITECTURE.md` in words rather than into a screen.

---
# chatcms-hq8m
title: A project duplicated, moved, zipped or archived
status: todo
type: epic
priority: normal
created_at: 2026-11-02T20:05:14Z
updated_at: 2026-11-02T20:05:14Z
parent: chatcms-7r0u
blocked_by:
    - chatcms-qllv
---

## Purpose

Four acts on a whole project that the design adds, each with its own rules
about what travels with it.

## Scope

- **Duplicate** (`org-projects-duplicate`): a separate copy in the same
  organisation. The dialog opens with the original's name, flagged invalid,
  because the copy must be given its own name. What is copied: pages, media
  and the content model always; people and their roles off by default;
  publishing settings off by default, "so the copy can't overwrite the live
  website". The dialog says what it does to the limit, and is blocked at it.
- **Move to another organisation** (`org-projects-move`): destination as
  radio cards. An organisation admin may choose only organisations they also
  administer; the rest are shown greyed with the reason ("You're a member
  here, not an admin"). A platform admin may choose any. The dialog says what
  happens: the website, publishing and history move along unchanged; the
  people on the project keep their roles and become members of the
  destination; the organisation admins of the old organisation lose sight of
  it. Blocked when the destination's member limit would be exceeded by the
  people who would come with it.
- **Download as zip**: pages, media and the content model, no dialog.
- **Archive**: the project stops counting toward the project limit and moves
  into the collapsed "Archived" group. Nothing is deleted.

## Done when

- Duplicating gives a project that shares nothing with its original: editing
  the copy leaves the original alone, proven against a real git remote.
- A copy never publishes to where the original publishes, unless somebody
  deliberately ticked publishing settings.
- Moving leaves the site in git untouched and carries the people with it,
  proven by an integration test over both organisations' membership.
- The zip holds what the design says and nothing else.
- Every one of the four is refused for somebody who may not do it, by
  `policy.Can`.
- The acceptance suite covers duplicating with a new name, moving to an
  organisation the person administers, and both blocked states.

## References

- `docs/design/DESIGN.md` 5.13; the four dialogs in
  `docs/design/reference/html/`.
- `internal/gitsync`, which owns everything that touches git.

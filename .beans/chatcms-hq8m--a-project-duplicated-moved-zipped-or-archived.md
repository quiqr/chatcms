---
# chatcms-hq8m
title: A project duplicated, moved, zipped or archived
status: completed
type: epic
priority: normal
created_at: 2026-11-02T20:05:14Z
updated_at: 2026-11-03T14:30:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-03-add-project-duplication-and-moving
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

## Summary of Changes

Shipped as
`openspec/changes/archive/2026-11-03-add-project-duplication-and-moving`.

### What a whole site can now be put through

- **Duplicated.** `internal/wholesite` copies the files of one site into a new
  one, with `.git` hidden from the copy, so a duplicate carries the pages, the
  media and the content model but none of the original's history and no
  remote. A copy takes a name of its own: the same name is refused with
  `refused.copy-needs-its-own-name`. People and publishing settings travel
  only when they are asked for, and the copy is taken under the same room
  check as a new site, with the row removed again when the files fail.
- **Moved.** `tenancy.MoveTo` moves the database row alone: the files and the
  history are untouched. Its people come with it and become members of the
  organisation it moves into, counted against that organisation's seats before
  anything moves. `policy.MayMoveTo` needs both ends, because taking a site
  out is the first organisation's to allow and putting one in is the second's.
  Whoever administers the installation may move a site anywhere, which is the
  one act over content that administering carries.
- **Downloaded.** A zip of the site's files, gated on seeing technical detail,
  buffered whole before a byte is written so a refusal is never half a file.
- **Archived.** A site out of the way stops counting against the
  organisation's limit of sites, and comes back only when there is room for it
  again.

### The screens

The four menu items in the organisation's site list stopped saying "Not built
yet": two open the dialogs the design draws, one is a plain download link so
the browser saves the file rather than the screen holding it, and one archives
in place. Archived sites sit in their own group below, which says it does not
count toward the limit.

### What the gate caught

The enforcement test `TestOnlyGitsyncRunsGit` failed on the new integration
test, which ran `git remote` to prove a copy publishes nowhere. The check is
right: nothing outside `internal/gitsync` runs git. The test reads `.git/config`
for a `[remote ` section instead, which proves the same thing without a second
place knowing how git is invoked.

### Proved by

Seven integration tests against a real git remote on disk, handler tests for
the four endpoints, six frontend tests over the dialogs and the archived
group, and `test/e2e/26-a-whole-site.spec.mjs` at 1440 px and at 390 px:
duplicated with a name of its own, moved to a second organisation, archived
and brought back, and downloaded as a zip that begins with `PK`.

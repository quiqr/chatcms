---
# chatcms-503w
title: 'The platform admin area: every user account, and what may be done to one'
status: completed
type: epic
priority: high
created_at: 2026-11-02T20:05:43Z
updated_at: 2026-11-03T03:00:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-03-add-user-administration
blocked_by:
    - chatcms-civ6
---

## Purpose

The people screen we shipped lists everybody and offers two things: turning an
account off and acting as somebody. The design asks for a screen that manages
an account: its name and address, its password reset, which organisations it
is in, whether it may sign in at all, and removing it.

## Scope

- **Users** (5.16): search by name or email, Invite user, and filter chips
  (All, Platform admins, Invited, Blocked, Not in any organisation).
- **The table**: the person with a "Platform admin" ink badge where it
  applies, the organisations they are in with "(admin)" beside the ones they
  administer, a status badge (Active, Invited, Blocked), when they last
  signed in, and a menu.
- **The menu**: edit name and email, send a password reset, add to an
  organisation, make or unmake a platform admin, block or unblock signing in,
  and delete the account.
- **Three states rather than two.** Today an account is on or off. The design
  asks for Invited (made, never accepted), Active, and Blocked. "Turning an
  account off" becomes "Block sign-in" in the words a person reads.
- **Deleting an account**, which is not the same as blocking it: what the
  person wrote stays, attributed as it was, and the account goes.
- **Pagination** ("Showing 1–8 of 46"), because the list we have answers one
  page and says how many there are in total but offers no way to the rest.

## Done when

- Every act in the menu works, is refused for anybody who does not administer
  the installation, and is written into the administration record.
- A blocked account cannot sign in and its sessions stop, which is what
  turning one off already does.
- A deleted account leaves the history of every site it touched readable.
- Searching and the filter chips answer from the server, in one query per
  page.
- The acceptance suite covers inviting, blocking, unblocking, editing a name,
  and deleting.

## References

- `docs/design/reference/html/admin-users.html`.
- What exists now: the people list in `web/src/routes/AdministrationPage.tsx`
  and `internal/admin/service.go`.

## Summary of Changes

The people list became a screen that manages an account.

- **Three states, one new column: none.** Blocked is `disabled_at`, which
  already existed; invited is "has never set a password"; active is neither.
  A third column would be a second thing to keep in step with the first.
- **The menu**: edit the name and address, send a password reset, add to an
  organisation, make or unmake a platform admin, block or unblock, delete.
  Each is refused for anybody who does not administer the installation and
  written into the record.
- **An address another account holds is refused by name**, caught as a unique
  violation and said in words rather than as a constraint.
- **Deleting is not blocking.** Blocking is for somebody who has left;
  deleting is for an account that should never have existed. Deleting takes
  the account and its memberships, and leaves what the person wrote: a
  version in git carries the name and address it carried at the time, and is
  not this database's to rewrite. A test proves the site survives and the
  membership does not.
- **Nobody blocks or deletes their own account.** The menu disables it and
  the server refuses it, for the same reason the last platform admin cannot
  stand down.
- **A password reset is sent, never set.** The same link the product already
  sends; `chatcms-294f` settled that nobody types somebody else's password,
  and this does not revisit it.
- **The list is the server's answer**: search, five filters, and a page at a
  time asked for by the last address rather than by an offset, so that
  somebody made while a platform admin is reading cannot make a row appear
  twice or not at all.
- Proved by five handler tests, four integration tests, four frontend tests
  of the screen, and `test/e2e/25-every-account.spec.mjs` at 1440 px and at
  390 px.

### A decision worth knowing

The list says which organisations somebody is in, and not which sites they
reach. That is what the hand-off draws, and the sites stay on the person's own
screen, where the detail belongs. `test/e2e/19` moved to assert the new
division rather than being loosened.

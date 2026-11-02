---
# chatcms-503w
title: 'The platform admin area: every user account, and what may be done to one'
status: todo
type: epic
priority: high
created_at: 2026-11-02T20:05:43Z
updated_at: 2026-11-02T20:05:43Z
parent: chatcms-7r0u
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

## Why

The people screen lists everybody and offers two things: turning an account
off, and acting as somebody. Everything else an account needs doing to has to
be done with SQL: changing a name somebody typed wrong, sending a reset to
somebody who cannot get in, putting somebody into an organisation, and
deleting an account that should never have existed.

The hand-off draws a screen that manages an account, with the three states an
account is actually in rather than the two the database has.

Epic `chatcms-503w`.

Bean: `.beans/chatcms-503w--the-platform-admin-area-every-user-account-and-wha.md`.

## What Changes

- **Three states rather than two**: invited, active and blocked. "Turning an
  account off" becomes blocking in the words a person reads.
- **A menu per account**: edit the name and address, send a password reset,
  add to an organisation, make or unmake a platform admin, block or unblock,
  and delete.
- **Deleting an account**, which is not blocking it: the account goes and
  what the person wrote stays, attributed as it was.
- **Filter chips and pages**: all, platform admins, invited, blocked, and in
  no organisation, answered by the server, with a page at a time and a way
  to the rest.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tenancy/administration`: what an account is, what may be done to one, and
  how the list is asked for.

## Impact

- **New**: endpoints for editing an account, resetting a password, adding
  somebody to an organisation and deleting an account; filters and paging on
  the list.
- **Changed**: the Users screen is the design's rather than the old stacked
  page's; an account's state is three things.

## Assumptions

- **Invited is read, not stored.** Somebody who has never set a password has
  been invited and has not accepted; blocked is the column that already
  exists. A third column would be a second thing to keep in step with the
  first.

- **Deleting an account leaves what the person wrote.** Versions in git carry
  a name and an address, and are not the database's to rewrite. The account
  goes, its memberships go with it, and the history of every site still
  reads as it did. This is why deleting and blocking are both offered: one
  for somebody who has left, one for somebody who should never have been
  here.

- **Nobody deletes or blocks themselves.** A platform admin who locks
  themselves out leaves an installation that only SQL can repair, which is
  the same reason the last administrator cannot stand down.

- **A password reset is sent, never set.** The menu sends the same link the
  product already sends; no administrator ever types somebody else's
  password, which `chatcms-294f` settled and this does not revisit.

- **The page is a page, not an offset into a shifting list.** Asking for the
  next page asks for what comes after the last address on this one, so that
  somebody being made while a platform admin is reading does not make a row
  appear twice or not at all.

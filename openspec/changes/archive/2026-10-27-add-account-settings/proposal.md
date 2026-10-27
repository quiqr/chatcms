## Why

`web/src/routes/AccountPage.tsx` can change a name and a password.
Everything else about an account is missing: there is no picture, the email
address cannot be changed, and a person cannot see where they are signed in
or end a session they left somewhere.

Epic `chatcms-clq4`.

Bean: `.beans/chatcms-clq4--account-settings-a-picture-an-email-address-and-yo.md`.

## What Changes

- **A picture for a person**, uploaded, read back, and replaced. Today
  `Avatar` computes letters from a name and nothing else exists.
- **Changing an email address**, which means proving the new one: a message
  to it, a link, and the old address working until the new one is confirmed.
- **Where you are signed in**: every session this account holds, when each
  was last used, which one you are reading this on, and ending one without
  ending them all.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `identity/accounts`: a picture, an email address that can change, and the
  sessions an account holds.

## Impact

- **New**: a migration for the picture and for an email address waiting to be
  confirmed, endpoints for all three, and the screens.
- **Changed**: `Avatar` shows a picture when there is one; `internal/auth`
  gains the picture, the change of address, and reading the sessions.

## Assumptions

- **A person's picture goes in PostgreSQL, not in a site's repository and not
  on disk.** Not the repository: it belongs to the account, crosses
  organisations, and a customer's published site must never carry a picture
  of somebody from another tenant. Not the state directory either: that is
  per-instance, not backed up, and `docs/DEPLOYING.md` says what to back up
  without mentioning it, so a picture there would be lost on the first
  restore. The database is the one thing with a restore story, and a picture
  bounded to a few hundred kilobytes is not what makes a database slow.

- **One size, made on the way in.** A picture is shown at 32 px and 44 px and
  nowhere larger. Keeping the original would mean storing megabytes to draw
  thumbnails, so it is resized once when it arrives and the original is not
  kept. `golang.org/x/image` is already a dependency.

- **What a file claims to be is not evidence.** The same rule the media
  library already follows: the bytes are read to decide what they are, and a
  file that is not a picture is refused however it is named.

- **An email address changes only once it is proven.** Until the link in the
  message sent to the new address is followed, signing in still uses the old
  one. Anything else lets a typo lock somebody out of their own account.

- **Ending one session is not ending them all.** Signing out everywhere
  exists already; this is for the laptop you left at the office, and it must
  leave the screen you are using alone unless you pick it.

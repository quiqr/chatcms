---
# chatcms-clq4
title: 'Account settings: a picture, an email address, and your sessions'
status: completed
type: epic
priority: normal
created_at: 2026-10-27T13:06:17Z
openspec-link: openspec/changes/archive/2026-10-27-add-account-settings
updated_at: 2026-10-27T13:57:20Z
parent: chatcms-8vem
blocked_by:
    - chatcms-k1hf
---

## Purpose

`web/src/routes/AccountPage.tsx` can change a name and a password, which
`chatcms-c8em` added. Everything else about an account is missing: the email
address cannot be changed, there is no picture, and a person cannot see where
they are signed in or end a session they have left somewhere.

This is the account itself. How the application behaves for somebody is
`chatcms-6h9u`, preferences, kept apart on purpose: one is who you are, the
other is how you like things.

## Scope

- **A picture.** Uploading one, a sensible size, and serving it. Today
  `Avatar` computes letters from a name and nothing else exists.
- **Where it lives has to be decided first.** Content pictures go in the
  site's git repository, which is right for them: they are the customer's and
  they get published. A person's picture is neither. It belongs to the
  account, it crosses organisations, and it must never end up in a customer's
  repository. So it needs the first blob storage this project has, with its
  own serving path and its own place in the backup story, which
  `docs/DEPLOYING.md` currently does not cover for the database either.
- **Changing an email address**, which means proving the new one: a message
  to it, a link, and the old address kept working until the new one is
  confirmed. The address is how somebody signs in, so getting this wrong
  locks them out.
- **Where you are signed in**: the sessions this account has, when each was
  last used, and ending one without ending them all. The data is in
  `sessions`; nothing shows it.
- **Removing your own account**, or saying plainly that only an
  administrator can and who to ask.

## Done when

- A picture can be added, replaced and removed, shows everywhere an avatar
  shows, and survives a restart.
- No person's picture is ever written into a site's repository, asserted by a
  test.
- Changing an email address needs the new one proven, and the old one keeps
  working until it is.
- Somebody can see where they are signed in and end one of those sessions.
- Every refusal is covered by an integration test against a real database.

## References

- `openspec/changes/archive/2026-10-27-add-people`, the half that is built.
- `web/src/components/ui/avatar.tsx`, initials and nothing more.
- `docs/DEPLOYING.md`, which needs the backup story either way.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-27-add-account-settings`.

### Where a picture lives, which the bean said to decide first

In PostgreSQL. Not in a site's repository: it belongs to the account, it
crosses organisations, and a customer's published site must never carry a
picture of somebody from another tenant. Not in the state directory either:
that is per-instance and was not in the backup story, so a picture there
would be lost on the first restore. The database is the one thing with a
restore story, and `docs/DEPLOYING.md` now names the picture as the one
thing a restore brings back that git cannot.

The rule is enforced by structure rather than by remembering: a test walks
every package that writes into a site and fails if any of them so much as
names a person's picture.

- **One size, made on the way in.** It is shown at 32 px and 44 px and
  nowhere larger, so the middle is cut out and scaled to 128 px square and
  the original is not kept.
- **What a file claims to be is not evidence.** The same rule the media
  library follows, now in `internal/imagekind` so both use one
  implementation rather than two copies. SVG is refused for an account,
  because a drawing can carry script and this one is shown inside the
  application.
- **An address changes only once it is proven.** The message goes to the new
  address, and signing in uses the old one until the link in it is followed.
  A typo cannot lock anybody out.
- **Where you are signed in**, with the screen being read marked and no way
  to end itself; that is what signing out is for.

### One defect this turned up

A picture may be kept by the browser for a minute, which is right everywhere
it is shown beside a name and wrong on the screen that has just changed it:
removing a picture left the old one on screen. The avatar now knows when it
changed.

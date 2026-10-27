---
# chatcms-clq4
title: 'Account settings: a picture, an email address, and your sessions'
status: in-progress
type: epic
priority: normal
created_at: 2026-10-27T13:06:17Z
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

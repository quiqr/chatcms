---
# chatcms-k1hf
title: Signing out, and a way to yourself from every screen
status: todo
type: epic
priority: high
created_at: 2026-10-27T13:05:56Z
updated_at: 2026-10-27T13:05:56Z
parent: chatcms-8vem
---

## Purpose

There is no way to sign out. `POST /api/sign-out` works and
`web/src/lib/session.ts` exports `signOut`, and nothing in the interface
calls it: a session can only be ended by waiting for it to expire. On a
shared or borrowed machine that is the difference between leaving and
leaving your account open.

Reaching your own account has the same shape. The avatar links to `/account`
in `OrganisationShell`, which is the projects, help and administration
screens. `ProjectShell`, `PhoneShell` and `EditorShell` draw an avatar that
is not a link, so from inside a site, where somebody spends their whole day,
there is no way to yourself and no way out.

## Scope

- A person menu, in every shell rather than in one: who you are, your
  account, and signing out. The same component in `ProjectShell`,
  `PhoneShell`, `EditorShell` and `OrganisationShell`, so the three that lack
  it cannot lack it again.
- Signing out calls the endpoint, forgets the session, and lands on the
  sign-in screen saying you have signed out, rather than on a screen that
  refuses to load.
- Signing out everywhere else as well, for a session left open somewhere you
  no longer have. `Sessions.RevokeAll` already does it; nothing offers it.
- A session that has already gone sends you to sign in, which the 401
  handling added in `add-organisations` does, and signing out must not look
  like that failure.
- On a phone, where the bottom bar has no room: the menu opens from the
  person at the top.

## Done when

- Signing out is reachable from every screen, at 1440 px and at 390 px.
- After signing out, the session is dead: going back in the browser does not
  show a signed-in screen, and the old cookie is refused.
- Signing out everywhere ends every other session, proven against a real
  database.
- An acceptance test signs in, signs out, and cannot get back in without
  signing in again.

## References

- `internal/httpapi/router.go`, the endpoint that exists.
- `internal/auth/session.go`, `Revoke` and `RevokeAll`.
- `web/src/shells/`, the four shells.

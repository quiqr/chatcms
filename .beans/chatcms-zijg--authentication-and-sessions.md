---
# chatcms-zijg
title: Authentication and sessions
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-22-add-authentication
updated_at: 2026-10-22T14:56:54Z
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

Let people sign in. Built-in email and password, sessions in Postgres, designed
so OIDC can be added without touching callers.

## Scope

- `internal/auth`: password hashing with argon2id, login, logout, session
  issue, lookup and revoke. Sessions in Postgres with an expiry and a
  "keep me signed in on this device" long-lived variant.
- Identity source behind an interface with one implementation, so OIDC can be
  added later. Sessions must not assume a password ever existed.
- Invitation acceptance: the link in the invitation email sets the first
  password. Tokens are single use and expire.
- Password reset by email link.
- The sign-in screen from DESIGN.md section 5.1, including the hint about
  invitation links.
- Rate limiting on login and reset, and no user enumeration in responses.

## Done when

- Signing in, staying signed in, and signing out all work and are tested
  against a real database.
- An expired or reused invitation token is refused with a plain-language
  message.
- Session fixation, missing CSRF protection on state-changing requests, and
  cookie flags are covered by tests.
- The screen matches `docs/design/reference/screenshots/desktop/login.png`.

## References

- `docs/BRIEFING-M01.md` section 4.1.
- `docs/design/DESIGN.md` section 5.1.

## Summary of Changes

Shipped as OpenSpec change `add-authentication`, archived at `openspec/changes/archive/2026-10-22-add-authentication`. Two capabilities specified and built, `identity/authentication` and `identity/credentials`, with a requirement added to `platform/server-runtime`.

Passwords are argon2id with the parameters written into the stored value, so raising them later leaves old hashes verifiable. A session is 32 random bytes in a cookie and a SHA-256 of them in the database, proven by a test that reads the rows directly and looks for the cookie value. Signing in always issues a new session and revokes whatever the request arrived with, which closes session fixation.

State-changing requests carry a double-submitted token: a readable cookie repeated in a header. Reads need none. Each half of the rule is covered, including a request with a cookie and no header and one where the two do not match.

Nothing reveals whether an address is known. A wrong password and an unknown address give the same error, and an unknown address still pays for a hash so it does not come back faster. A reset request answers identically either way. A spent link, an expired link and one that never existed are one error.

Guessing is limited per address and per caller, so neither varying the address nor varying the caller helps, and a correct sign-in clears the count.

The identity seam is in place: `Identity.Authenticate` takes a `Claim`, and everything that reads a session works for a person with no password, which is what an invited person is before they accept.

The sign-in screen is built as the designs draw it, with the invitation hint. Checked in a browser at 1440 and 390: no sideways scrolling and nothing under forty-four pixels, after widening the forgot-password link.

Coverage: `internal/auth` 84.4%, `internal/httpapi` back above its floor at 89%, overall 86.5%.

Two things recorded rather than done. Nothing is emailed, because there is no mail service; a link goes to the server log and never to the browser, and the epic that adds mail replaces one function. And the designs decorative mini-form illustration beside the sign-in form is not drawn, since it illustrates a screen that does not exist yet.

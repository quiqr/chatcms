---
# chatcms-zijg
title: Authentication and sessions
status: in-progress
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T14:37:41Z
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

---
# chatcms-axh4
title: 'Mail: sending invitations and resets, with Mailpit in development'
status: in-progress
type: epic
priority: high
created_at: 2026-10-26T16:45:24Z
updated_at: 2026-10-26T21:48:31Z
parent: chatcms-8vem
---

## Purpose

Invitations and password resets are issued but never sent. Today the link is
written to the server log with a warning saying mail is not set up, which
means no customer can be invited without an implementer reading a log and
pasting a link by hand. `internal/auth/service.go` says in a comment that the
epic adding mail replaces that function; this is that epic.

## Scope

- `internal/mail`: a sender behind one interface, SMTP underneath, with the
  host, port, credentials and envelope address configurable the way everything
  else is, including from a file so a password stays out of the Nix store.
- The messages themselves: invitation, password reset, and the one that says
  a password was changed. Plain text and HTML, in the editor-facing language
  the copy rules ask for, with no git words in any of them.
- Sending happens through the job queue, not in the request, so a slow or
  refused mail server never makes somebody wait or lose an invitation. A
  failed send is retried and then visible as a failed job rather than
  forgotten.
- Mailpit in the dev shell, which is the maintained successor to MailHog:
  SMTP on 1025, a web inbox on 8025, nothing stored outside the state
  directory. `scripts/dev.sh` starts it with everything else and prints where
  to read the mail, and stops it on the way out.
- The NixOS module gains the same settings, and `docs/DEPLOYING.md` says what
  a real deployment needs.
- When no mail server is configured at all, the link is still logged, because
  a single-tenant install run by the person who owns it should not be forced
  to set up SMTP before anybody can sign in. That path says so plainly rather
  than looking like a failure.

## Done when

- Inviting somebody on the test instance puts a real message in Mailpit, and
  the link in it sets a password and signs that person in.
- A refused mail server leaves a failed job and a message an implementer can
  read, and never loses the invitation or breaks the screen that sent it.
- An acceptance test invites somebody, reads the message out of Mailpit over
  its API, and follows the link. No log scraping.
- `nix flake check` covers it, Mailpit included.

## References

- `docs/BRIEFING-M01.md` section 4, people and invitations.
- `internal/auth/service.go`, `recordLink`.
- `docs/design/DESIGN.md` section 7, the copy rules.

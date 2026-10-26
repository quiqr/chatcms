## Why

Invitations and password resets are issued and never sent. The link goes to
the server log with a warning saying mail is not set up, so nobody can be
invited without an implementer reading a log and pasting a link by hand.
`internal/auth/service.go` says in a comment that the epic adding mail
replaces that function. Epic `chatcms-axh4`.

While building it, two screens turned out to be missing: the link in an
invitation points at `/set-password`, and nothing is routed there. Sending a
message whose link lands on a not-found page would be worse than not sending
it, so they are part of this change.

Bean: `.beans/chatcms-axh4--mail-sending-invitations-and-resets-with-mailpit-i.md`.

## What Changes

- **`internal/mail`**: a sender behind one interface, SMTP underneath, with
  everything configurable the way the rest is, including from a file so a
  password stays out of the Nix store.
- **Three messages**, in the language the copy rules ask for: an invitation, a
  password reset, and word that a password was changed. Plain text and HTML.
- **Sending goes through the job queue**, never in the request, so nobody
  waits on somebody else's mail server and a slow one cannot lose an
  invitation. A failure is retried; having nowhere to send at all is not.
- **Mailpit in the development shell**, started and stopped by
  `scripts/dev.sh`, with its inbox named in the banner.
- **The two screens the links need**: choosing a password from an invitation
  or a reset, and asking for a reset.
- **With no mail server configured, the link still goes to the log**, because
  an install run by the person who owns it should not have to set up SMTP
  before anybody can sign in.

## Capabilities

### New Capabilities

- `platform/mail`: what is sent, how it is sent, what happens when it cannot
  be, and what a message may never contain.

### Modified Capabilities

- `identity/credentials`: an invitation and a reset are now sent, not only
  issued.

## Impact

- **New**: `internal/mail`, `internal/app/mail.go`,
  `web/src/routes/SetPasswordPage.tsx`,
  `web/src/routes/ForgotPasswordPage.tsx`, `test/e2e/10-invitations.spec.mjs`.
- **Changed**: `internal/auth` sends instead of only logging; the harness
  runs a Mailpit of its own so a test reads the message rather than the log;
  the NixOS module and `scripts/dev.sh` carry the settings.

## Assumptions

- **Mailpit, not MailHog.** MailHog has had no release since 2020. Mailpit is
  the maintained successor with the same shape: SMTP on one port, a web inbox
  on another, nothing kept outside the state directory.

- **Sending belongs on the queue.** In the request it would make somebody wait
  on a server that is not ours, and a refusal would lose the invitation along
  with it. On the queue a failure is retried and then visible as a failed job.

- **Nowhere to send is not a failure.** A single-tenant install run by its
  owner should not be made to configure SMTP before anybody can sign in. The
  link goes to the log, and the log says plainly that is what happened.

- **A message that cannot be read is dropped, not retried.** A payload that
  will not parse will not parse in five minutes either, and a job that can
  never succeed buries the ones that would.

- **A message never carries a git word.** The copy rules apply to what is
  sent as much as to what is shown, and a test asserts it on the invitation,
  because an invitation is the first thing a content creator ever reads.

- **A reset says the same thing whatever the answer.** Telling somebody that
  an address is unknown would turn the screen into a way of finding out who
  has an account.

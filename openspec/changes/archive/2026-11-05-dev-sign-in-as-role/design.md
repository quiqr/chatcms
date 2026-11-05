# Design

## Context

See proposal.md for the motivation. The seed (`cmd/seed`) creates
`owner@example.org` and `admin@example.org` with a passphrase, and invites
`marieke@korenaar.example` and `sam@korenaar.example` through
`tenancy.Service.InvitePerson`, which leaves them without one. The sign-in
screen (`web/src/routes/SignInPage.tsx`) holds email and password in state and
submits them to the existing password sign-in.

## Goals / Non-Goals

**Goals:**

- One click per role under `npm run dev`.
- Nothing of it in a production bundle, enforced by the gate.

**Non-Goals:**

- No server endpoint that signs in without a password, behind a flag or not.
- No buttons in a `nix build` binary run locally.
- No extra roles beyond the four the seed already has.

## Decisions

### Gate on `import.meta.env.DEV` in the frontend, not on a server flag

Vite replaces `import.meta.env.DEV` with `false` in a production build, so a
branch behind it, and a module only imported from it, is left out of the
bundle. The alternative, a server endpoint behind `CHATCMS_DEV`, would put a
passwordless sign-in in every binary, one forgotten setting away from a
deployment.

### The buttons submit the ordinary sign-in

A button sets email and password and calls the same submit the form uses.
Every refusal, the session, and staying signed in behave as they do for a
typed sign-in, and nothing new sits behind the identity interface.

### The development block lives in its own module

`DevSignIn.tsx` holds the four people and the passphrase, and is reached only
through a dynamic import inside `if (import.meta.env.DEV)`. That keeps the
data out of `SignInPage.tsx` and makes "not in the bundle" true by
construction rather than by minifier luck.

### The seed gives the invited people a passphrase directly

After `InvitePerson`, the seed calls `People().SetPassword` with the returned
person id. Going through `SetPasswordWithToken` would mean fishing the token
out of the mail the seed cannot send. A membership reads as still to accept
only while the person has no password, so setting one is all accepting takes.

### The bundle check is a grep over `dist`

A script run by the `frontend` check fails when the built assets contain the
seeded passphrase or one of the seeded addresses. This is cheap and catches a
refactor that moves the data out of the guarded import. The block's labels
are not checked for, because the catalogue is bundled whole.

## Risks / Trade-offs

- [The seeded passphrase sits in the repository] → It already does in
  `cmd/seed` and `scripts/dev.sh`; it only works against a seeded development
  database.
- [Someone runs `npm run dev` against a real server] → The buttons only try
  the seeded addresses with the seeded passphrase, which a real installation
  does not have; they fail like a wrong password.
- [`check:strings` flags the addresses] → They are data in the development
  module, not text shown as a string literal in JSX; labels reuse the
  catalogue's existing role names.

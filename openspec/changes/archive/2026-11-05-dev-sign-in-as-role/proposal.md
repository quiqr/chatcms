# Proposal

## Why

Trying a screen as each role now means typing an address and a passphrase,
and only two of the four seeded people can sign in at all: the editor and the
viewer on `korenaar` are invited but never accepted. Checking what an editor or
a viewer sees is the most common thing done by hand during development, so it
should take one click.

Originating bean: none. This came up while running the test instance on a
development machine and is tooling for development, not part of a milestone
epic.

Acceptance test served: none directly. It makes the hand-checking behind the
role-related tests in `docs/BRIEFING-M01.md` section 6 quicker; it adds no
acceptance test of its own.

## What Changes

- The seed accepts the invitations of `marieke@korenaar.example` (editor) and
  `sam@korenaar.example` (viewer) with the same passphrase as the other seeded
  people, so all four roles can sign in.
- The seed invites `noor@korenaar.example` (editor) and leaves that invitation
  open, so an open invitation stays visible; acceptance test 19 looks at her
  instead of Marieke.
- The sign-in screen, when served by the Vite development server, shows a
  block marked as development with one button per role: administrator,
  implementer, editor, viewer. A button fills in that person's address and
  passphrase and submits the ordinary sign-in form.
- A production build of the frontend contains neither the block nor the
  seeded passphrase.
- `dev.sh` and the seed print the editor and viewer alongside the other two.

## Capabilities

### New Capabilities

- `platform/local-development`: what a development machine gets that a
  deployment never does: the seeded people and projects, and signing in as one
  of them in one step.

### Modified Capabilities

None. Signing in itself does not change; the buttons go through the existing
password sign-in.

## Impact

- `cmd/seed/main.go`: give the two invited people a passphrase.
- `scripts/dev.sh`: the notice lists all four people.
- `web/src/routes/SignInPage.tsx` and a new development-only component.
- `web/src/strings/en.ts` and every other catalogue: labels for the block and
  its buttons.
- A check on the production bundle, run by the `frontend` check in
  `flake.nix`.
- Seam: auth. Nothing new sits behind the identity interface; there is no
  sign-in without a password and no server endpoint is added.

### Assumptions

- `korenaar` is enough for the editor and the viewer; `studio-noord` stays a
  project with only the implementer.
- The buttons appear only under `npm run dev`, not in a binary built with
  `nix build`, even when that binary runs on a development machine.
- The seeded addresses and passphrase are development data, not interface
  text, so they stay out of the string catalogue.

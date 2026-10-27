## Why

There is no way to sign out. `POST /api/sign-out` works and
`web/src/lib/session.ts` exports `signOut`, and nothing in the interface
calls it: a session can only be ended by waiting for it to expire. On a
borrowed or shared machine that is the difference between leaving and leaving
your account open.

Reaching your own account has the same shape. The avatar links to `/account`
in `OrganisationShell`; `ProjectShell`, `PhoneShell` and `EditorShell` draw
one that is not a link. From inside a site, where somebody spends the whole
day, there is no way to yourself and no way out.

Epic `chatcms-k1hf`.

Bean: `.beans/chatcms-k1hf--signing-out-and-a-way-to-yourself-from-every-scree.md`.

## What Changes

- **One person menu, in every shell**: who you are, your account, and signing
  out. The same component in all four, so three of them cannot lack it again.
- **Signing out ends the session and says so**, landing on the sign-in screen
  with a word about what happened rather than on a screen that refuses to
  load.
- **Signing out everywhere else**, for a session left open somewhere you no
  longer have. `Sessions.RevokeAll` exists and nothing offers it.
- **A test that the shells cannot drift**, because this defect was three
  shells quietly missing what the fourth had.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `identity/authentication`: signing out ends a session, and can end every
  other one.

## Impact

- **New**: `web/src/shells/PersonMenu.tsx`, an endpoint for signing out
  everywhere.
- **Changed**: the four shells use the menu; `internal/auth` gains signing
  out everywhere.

## Assumptions

- **Signing out is not the same as a session that has run out.** Both end on
  the sign-in screen, and they must not look alike: one is something somebody
  did, the other is something that happened to them. Signing out says so.

- **Signing out everywhere does not spare the session asking.** Somebody who
  thinks their account is not theirs any more wants every door shut,
  including the one they are standing in. They are signed out here too.

- **The menu is one component, not four copies.** This defect existed because
  one shell had the link and three did not. A shared component plus a test
  that walks all four is what stops it happening again.

- **On a phone the menu opens from the person at the top**, because the
  bottom bar is the site's sections and has no room for an account.

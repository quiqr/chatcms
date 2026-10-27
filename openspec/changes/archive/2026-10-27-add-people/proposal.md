## Why

People can be invited to a project and given a role there. Nothing else about
a person can be managed: not at organisation level, not by the person
themselves, and not by whoever has to turn an account off when somebody
leaves. `organisation_members` has held roles since the tenancy epic and
until the organisations epic no endpoint wrote to it.

Epic `chatcms-c8em`.

Bean: `.beans/chatcms-c8em--people-organisation-membership-turning-accounts-of.md`.

## What Changes

- **Turning an account off**, and on again. Somebody turned off cannot sign
  in and their sessions stop at once.
- **What they wrote stays theirs.** Turning an account off changes nothing
  about the versions they made.
- **A person's own account**: their name and their password, changed with
  the one they have.

## Capabilities

### New Capabilities

- `identity/accounts`: what a person may change about themselves, and what
  turning an account off means.

### Modified Capabilities

None.

## Impact

- **New**: endpoints for an account and for turning one off,
  `web/src/routes/AccountPage.tsx`, and the administration controls for
  somebody who has left.
- **Changed**: `internal/auth` can turn an account off, rename somebody and
  change their password; `internal/admin` offers it and writes it down.

## What this change does not do

The bean also asks for an organisation's own owner to invite people to it,
for resending and withdrawing an invitation, and for the project and
organisation switch on the people screen. None of that is here. An
administrator can already put somebody into an organisation and take them
out, which is enough for a customer to be set up; the owner doing it for
themselves is a change of its own rather than something half-built into this
one. The bean says so and stays open.

## Assumptions

- **Turning an account off is not deleting it.** What somebody wrote stays
  attributed to them: history that renames its author is history nobody can
  trust, and a customer looking at last year's changes should see who really
  made them.

- **Sessions end at once.** An account turned off whose session keeps working
  until it expires is an account that is not really off. The sessions are
  revoked in the same step.

- **Nobody can turn their own account off.** It cannot be undone by the
  person who did it, and an organisation whose only owner locked themselves
  out needs somebody with a database client.

- **Changing a password needs the current one**, even while signed in,
  because a session left open on somebody else's machine should not be enough
  to take an account over.

- **Turning an account off belongs to the administrator**, not to an
  organisation's owner. An account is held by the installation: somebody who
  belongs to two organisations would otherwise be locked out of one by the
  other.

- **An email address is not changed in this milestone.** Changing it means
  proving the new one, which means another message and another link, and the
  briefing does not ask for it. The screen says so rather than offering a
  field that half works.

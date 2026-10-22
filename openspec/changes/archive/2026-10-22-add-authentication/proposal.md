## Why

Nothing in the product can be anyone's until the product knows who anyone is.
Every capability shipped so far takes an anonymous actor from the request
context; this change puts a person there. Epic `chatcms-zijg`.

It serves acceptance test 1, "create an organisation and invite the customer
as editor", together with the tenancy epic, and it is the precondition for
every screen that shows one person's work rather than everyone's.

## What Changes

- **Someone can sign in.** Email and password, with the password stored as an
  argon2id hash and never as anything that can be turned back into a password.
- **Signing in produces a session, not a token in a page.** Sessions live in
  PostgreSQL, are addressed by a cookie that JavaScript cannot read, and are
  stored as a hash so a copy of the database does not hand out sessions.
- **Staying signed in is a choice.** "Keep me signed in on this device" issues
  a long-lived session; without it the session lasts a working day.
- **A session does not assume a password ever existed.** The thing that
  proves who someone is sits behind an interface with one implementation, so
  single sign-on can be added without touching anything that reads a session.
- **An invitation sets the first password.** The link in an invitation is
  single use, expires, and is stored as a hash. So is a password reset link.
- **Guessing is made expensive and unrewarding.** Sign-in and reset are rate
  limited per address and per caller, and neither ever reveals whether an
  address is known.
- **State-changing requests carry proof they were meant.** A request that
  changes anything must present a token that only this site's own pages can
  read, so another site cannot make a reader's browser act for them.
- **The sign-in screen exists**, as the designs draw it, including the line
  about using the invitation link first.

## Capabilities

### New Capabilities

- `identity/authentication`: how someone proves who they are, how that is
  remembered, and how it stops being remembered.
- `identity/credentials`: how a password is set, changed and checked, and how
  an invitation or a reset is issued and spent.

### Modified Capabilities

- `platform/server-runtime`: a request now carries the person who made it, and
  a request that changes anything has to prove it was meant.

## Impact

- **New code**: `internal/auth`, handlers in `internal/httpapi`, the sign-in
  screen in `web/`.
- **Changed code**: `internal/httpapi` replaces its single anonymous actor
  with one taken from a session; `internal/app` wires the session store.
- **Schema**: one migration adding people, sessions and the single-use tokens
  behind invitations and resets.
- **Dependencies added**: `golang.org/x/crypto` for argon2id. Nothing else.
- **Seams**: this change implements the fourth seam named in
  `docs/ARCHITECTURE.md`. Everything that reads a session goes through one
  function, and the thing that checks a password is one implementation of an
  interface.

## Assumptions

- **Nothing is emailed yet.** There is no mail service, and inventing one here
  would be guessing at a deployment decision. An invitation or a reset records
  its link in the server log, where an implementer can read it, and the API
  returns nothing that would let a caller harvest links. The epic that adds
  mail replaces the one function that delivers.
- **Rate limiting is in this process.** A single server is the deployment
  model, so the limiter is in memory and resets when the server restarts. The
  design records what changes when there is more than one server.
- **Permissions are not here.** This change establishes who someone is, not
  what they may do. `internal/policy` and the roles are the tenancy epic's,
  and until it lands every signed-in person can reach every signed-in
  endpoint, of which there are two.
- **No password rules beyond length.** A minimum length and a check against
  the most common passwords, and no composition rules, because composition
  rules produce worse passwords. The number is recorded in the design.

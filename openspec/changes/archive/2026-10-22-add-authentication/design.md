## Context

`internal/httpapi` already puts an actor in the request context and hands it
to handlers; it has been a single anonymous one since the foundation change.
This change makes it a person. See `proposal.md` for why it comes before
tenancy.

## Goals / Non-Goals

**Goals:**

- A stolen database does not yield usable sessions or recoverable passwords.
- Nothing anywhere tells an attacker whether an address is known.
- Adding single sign-on later touches one file.

**Non-Goals:**

- Permissions. Who may do what is the tenancy epic's.
- Sending mail. See the assumption in `proposal.md`.
- Multi-factor authentication, which milestone 1 does not call for and which
  would be better added once there is an identity interface to add it behind.

## Decisions

### argon2id, with parameters recorded in the hash

Passwords are hashed with argon2id at 64 MiB of memory, one pass and four
lanes, with a 16-byte salt and a 32-byte key. The encoded form carries the
algorithm and the parameters, so raising them later does not invalidate
anything: an old hash still verifies with the parameters it was made with.

Alternative considered: bcrypt, which is in the standard-adjacent libraries
and simpler. Rejected because argon2id is what current guidance asks for and
the cost of choosing it now is one dependency that is already in the module
graph.

### A session is a random value; the database holds its hash

The cookie carries 32 random bytes; the row holds a SHA-256 of them. Looking
up a session hashes the presented value and finds the row. A plain hash is
right here, unlike for passwords: the value has full entropy, so there is
nothing to guess, and lookups must be fast.

This is the decision that makes a database copy useless for impersonation,
which is worth more than it costs.

### Sign-in always issues a new session and kills the one it arrived with

Session fixation is the attack where someone hands a victim a session value
and waits for them to sign in with it. Issuing a new session on every sign-in
and revoking whatever arrived closes it. It is two lines and it is the kind of
thing that is never added later.

### Cross-site request forgery: a double-submitted token

A cookie that scripts can read carries a random token; a state-changing
request must repeat it in a header. Another site can make a browser send the
session cookie, but cannot read the token cookie to repeat it, and cannot set
the header without the server's permission.

`SameSite=Lax` on the session cookie is the first line and would nearly
suffice on its own. The second line exists because "nearly" is doing a lot of
work in that sentence, and because the token is five lines of middleware.

Reads need no token: a request that changes nothing cannot be abused this way,
and requiring it would make every page load carry ceremony.

### Rate limiting is per address and per caller, in memory

Two buckets: one keyed by the address being attempted, one by the caller's
address. Limiting only by address lets one caller work through a list of
people; limiting only by caller lets a botnet work on one person. Both are
cheap.

In memory means it resets on restart and does not span servers. The deployment
model is one server, so this is correct today and is called out here because
the day there are two, it silently becomes weaker rather than failing.

### Saying nothing, consistently

An unknown address and a wrong password give the same answer, and the same
timing: when the address is unknown, a hash is still computed against a dummy
value so the response does not come back faster. Asking for a reset answers
the same for a known and an unknown address.

The sign-in screen's hint about invitations exists precisely so that someone
who cannot get in has somewhere to go that does not require the system to
confirm anything.

### Tokens for invitations and resets are one mechanism

Both are a row with a kind, a hash of the token, an expiry and a spent-at. One
table, one issue function, one spend function, two kinds. The difference is
how long they last and what the message says.

Spending a reset revokes every session that person has, because a reset is
what someone does when they think their account is not theirs any more.

### Passwords: a length floor, a common-password list, no composition rules

Twelve characters, and a refusal of the few thousand most common passwords.
No rules about digits or symbols, because such rules reliably produce
`Password1!` and nothing better.

The list ships in the binary rather than being fetched; it is small.

## Risks / Trade-offs

- **In-memory rate limiting resets on restart**, so a determined attacker can
  restart-time their guessing if they can cause restarts. They cannot, and the
  window is small. The day there is a second server this moves to the database.
- **argon2id at 64 MiB costs memory per sign-in.** A burst of sign-ins is a
  burst of memory. Mitigated by the rate limiter, which caps the rate before
  it caps the memory.
- **Logging an invitation link puts a credential in the log.** It is the least
  bad option while there is no mail, it is stated in the proposal, and the log
  is readable only by whoever runs the server. The epic that adds mail removes
  it.
- **A long-lived session is a long-lived risk.** Mitigated by revoking every
  session on a password reset, and by the session being revocable server-side
  at any time.

## Migration Plan

One migration adds people, sessions and tokens. Nothing is deployed, so there
is nothing to migrate. The first person is created by an implementer through a
command the deployment epic adds; until then, tests create people directly.

## Open Questions

- Whether a session should be extended as it is used, rather than expiring at
  a fixed time from issue. Fixed expiry is simpler and safer; sliding expiry
  is friendlier. It can be changed without changing anything that reads a
  session.

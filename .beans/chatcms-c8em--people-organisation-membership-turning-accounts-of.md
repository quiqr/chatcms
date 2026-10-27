---
# chatcms-c8em
title: 'People: organisation membership, turning accounts off, and your own account'
status: in-progress
type: epic
priority: high
created_at: 2026-10-26T16:46:03Z
openspec-link: openspec/changes/archive/2026-10-27-add-people
updated_at: 2026-10-27T00:40:05Z
parent: chatcms-8vem
blocked_by:
    - chatcms-8n4m
    - chatcms-rbdb
---

## Purpose

People can be invited to a project and given a role there. Nothing else about
a person can be managed: not at organisation level, not by the person
themselves, and not by whoever has to turn an account off when somebody
leaves. `organisation_members` holds roles already; no endpoint reads or
writes it.

## Scope

- Organisation people: invite, change the organisation role, remove. The
  "This project / Organisation" switch the design draws on the people screen,
  which is listed as missing in `docs/design/COMPARISON.md`.
- Turning an account off and on again. A person who is off cannot sign in and
  their sessions end, but what they wrote stays attributed to them, because
  history that renames its author is history nobody can trust.
- Resending an invitation that was never accepted, and withdrawing one.
- A person's own account: their name, their email address, and changing their
  password with the current one.
- "Last active", which the people screen draws and does not have. Sessions
  already record when they were last seen.
- Roles stay the four the policy knows. This epic does not add a role; it
  gives the existing ones somewhere to be set.

## Done when

- An owner can invite to the organisation, change a role, turn an account off,
  and turn it on again, on desktop and at 390 px.
- A person turned off cannot sign in, and any session they had stops working
  at once rather than at its own expiry.
- Their name stays on the versions they made.
- Somebody can change their own password, and cannot change anybody else's.
- An integration test covers every one of those against a real database, and
  `policy.Can` is the only thing deciding who may.

## References

- `docs/BRIEFING-M01.md` section 4, people and roles.
- `docs/design/reference/screenshots/desktop/people.png`.
- `docs/design/COMPARISON.md`, the People row.

## Partly shipped

`openspec/changes/archive/2026-10-27-add-people` covers the account half.

### Done

- **An account can be turned off**, and on again. Somebody turned off cannot
  sign in, and whatever they had open stops working at once rather than at
  its own expiry.
- **What they wrote stays theirs.** Turning an account off changes nothing
  about the versions they made, and a test says so.
- **Nobody turns their own off**, because it cannot be undone by the person
  who did it.
- **Only an administrator may.** An account is held by the installation, so
  an organisation's owner turning one off would lock that person out of every
  other organisation too. There is a test for that refusal.
- **A person's own name and password**, with the password they have asked for
  even while signed in: a screen left open on somebody else's machine should
  not be enough to take an account over.

### Not done, and why

- An organisation's own owner inviting people to it, changing what they hold
  there, and taking them out. An administrator can already do all three,
  which is enough to set a customer up, so this is a change of its own rather
  than something half-built into that one.
- Resending an invitation that was never accepted, and withdrawing one.
- The project and organisation switch on the people screen, which waits on
  the above.
- Last active. The server can answer it; no screen asks yet.

Those four are `chatcms-9lu6`.

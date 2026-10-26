---
# chatcms-c8em
title: 'People: organisation membership, turning accounts off, and your own account'
status: todo
type: epic
priority: high
created_at: 2026-10-26T16:46:03Z
updated_at: 2026-10-26T21:01:51Z
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

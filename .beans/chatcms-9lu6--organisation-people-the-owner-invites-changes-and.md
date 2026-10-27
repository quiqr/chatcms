---
# chatcms-9lu6
title: 'Organisation people: the owner invites, changes and removes'
status: in-progress
type: epic
priority: high
created_at: 2026-10-27T01:05:55Z
updated_at: 2026-10-27T13:32:13Z
parent: chatcms-8vem
---

## Purpose

`chatcms-c8em` shipped the account half: an account can be turned off, and a
person manages their own name and password. The organisation half is still
missing, and an administrator doing it on the owner's behalf is not the same
as the owner doing it.

## Scope

- An organisation's owner invites somebody to the organisation, changes what
  they hold there, and takes them out. The same policy question an
  administrator goes through, asked of the owner.
- Resending an invitation that was never accepted, and withdrawing one.
  Withdrawing takes that person out.
- The "This project / Organisation" switch the design draws on the people
  screen, listed as missing in `docs/design/COMPARISON.md`.
- Last active, which `People.LastSeen` already answers and no screen asks
  for.

## Done when

- An organisation's owner can invite, change a role and remove, on desktop
  and at 390 px, and somebody who merely belongs cannot.
- A change that would leave an organisation with no owner is refused, with a
  message saying why.
- An invitation can be sent again and withdrawn, and withdrawing takes the
  person out.
- The people screen shows both levels, and when each person was last here.
- Integration tests cover every refusal against a real database.

## References

- `.beans/chatcms-c8em--people-organisation-membership-turning-accounts-of.md`
- `openspec/changes/archive/2026-10-27-add-people`
- `docs/design/reference/screenshots/desktop/people.png`

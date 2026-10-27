---
# chatcms-d4f3
title: 'Administration: open, change and remove an organisation'
status: in-progress
type: epic
priority: high
created_at: 2026-10-27T15:17:16Z
updated_at: 2026-10-27T18:33:22Z
parent: chatcms-8vem
---

## Purpose

The administration screen lists every organisation and offers nothing to do
with one. Renaming exists and is an administrator's, but the only way in is
to know the address of that organisation's own settings screen. Removing one
is a rule `chatcms-rbdb` states and no screen offers.

## Scope

- Each organisation on the administration screen opens: its name, its
  address, who is in it and what they hold there, and its projects.
- Renaming from there, which already works and is only unreachable.
- Removing an organisation, which the policy allows and nothing offers. The
  confirmation names what goes: its projects stop being reachable through
  ChatCMS, and every site in git is untouched.
- Putting somebody in and taking them out, which exists over the API from
  `chatcms-rbdb` and has no screen.
- What this does not add: a second way to do what an organisation's owner
  already does for themselves. The administrator's screen is for an
  organisation with nobody left to run it, or one being set up.

## Done when

- Every organisation can be opened, renamed, given an owner and removed from
  the administration screen, at 1440 px and at 390 px.
- Removing one leaves every git remote exactly as it was, proven by a test.
- An organisation with no owner left can be given one.
- Every act is in the administration record, as `chatcms-rbdb` requires.

## References

- `openspec/changes/archive/2026-10-26-add-application-administrator`
- `openspec/changes/archive/2026-10-27-add-organisations`

---
# chatcms-d4f3
title: 'Administration: open, change and remove an organisation'
status: completed
type: epic
priority: high
created_at: 2026-10-27T15:17:16Z
updated_at: 2026-10-27T19:40:00Z
parent: chatcms-8vem
openspec-link: openspec/changes/archive/2026-10-27-add-organisation-administration
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

## Summary of Changes

An organisation opened from the administration screen now shows everything in
it at once: its name, who is in it with what role, and the sites it holds.

- `GET /api/administration/organisations/{organisation}` answers with the
  organisation, its people and its sites in one view; always as lists, so an
  empty organisation reads as nothing rather than as nothing at all.
- `DELETE` on the same address removes one. It is refused while the
  organisation still holds sites, and the refusal names them, so that letting
  go of several customers' work is never one confirmation.
- Removing touches the database only. A test makes a real remote through
  `internal/gitsync`, removes the site and then the organisation, and proves
  the remote is still there afterwards.
- The screen renames, puts somebody in, takes somebody out, and removes after
  a confirmation that says in plain words what stays.
- Proved by handler tests for both endpoints, five integration tests against a
  real database, and `test/e2e/18-administering-an-organisation.spec.mjs` at
  1440 px and at 390 px.

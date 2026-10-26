---
# chatcms-8n4m
title: 'Organisations: make, rename and switch, instead of one hardcoded slug'
status: todo
type: epic
priority: high
created_at: 2026-10-26T16:45:44Z
updated_at: 2026-10-26T21:01:51Z
parent: chatcms-8vem
blocked_by:
    - chatcms-rbdb
---

## Purpose

An organisation can be created over the API and listed, but there is no way
to make one, rename one, or see who is in it from the screen. The frontend
also still hardcodes the organisation: every project route in
`web/src/App.tsx` passes `organisation="technative"`, which the seed happens
to create. A second organisation would therefore open the wrong one.

## Scope

- The organisation comes from the address, not from a constant. `App.tsx`
  stops passing a hardcoded slug, routes carry the organisation, and the
  switcher in the top bar changes which one you are looking at.
- Making an organisation from the switcher, the way the drawing shows it:
  name, the address it will get suggested from the name, and then its empty
  projects screen. Only an administrator sees that entry, and only an
  administrator may use it; `chatcms-rbdb` adds the role and the rule, and
  this epic is the screen for it.
- An organisation screen: its name, its address, the people in it with their
  organisation role, and renaming. Deleting is out of scope for this
  milestone and says so on the screen rather than being absent.
- Who may do what comes from `policy.Can`, not from new checks beside it. An
  administrator makes and renames an organisation; its owner manages the
  people in it; everybody else sees it and nothing more. Removing one is a
  rule `chatcms-rbdb` states and no screen here offers.
- A person who belongs to no organisation sees a screen that says so, rather
  than an empty list under a placeholder name. It offers to make one only to
  an administrator; anybody else is told to ask whoever invited them.

## Done when

- Two organisations exist on the test instance and switching between them
  shows each one's own projects, with nothing leaking between them.
- An administrator makes an organisation from the screen and lands on its
  projects; somebody who is not an administrator has no such entry and is
  refused by the policy if they ask for it anyway.
- Renaming shows everywhere the name is shown, without a reload.
- An integration test proves somebody in one organisation cannot read the
  other's projects through the API, whatever the address says.
- No file under `web/src/` names an organisation.

## References

- `docs/BRIEFING-M01.md` section 3, the two audiences.
- `docs/design/reference/screenshots/desktop/projects.png`, the switcher.
- `internal/httpapi/router.go`, the organisation endpoints that exist.
- `docs/design/COMPARISON.md`, the projects screen row.
- `chatcms-rbdb`, which adds the administrator this epic draws the screens
  for.

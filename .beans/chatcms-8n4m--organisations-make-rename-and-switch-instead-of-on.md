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
  projects screen.
- An organisation screen: its name, its address, the people in it with their
  organisation role, and renaming. Deleting is out of scope for this
  milestone and says so on the screen rather than being absent.
- Who may do what comes from `policy.Can`, not from new checks beside it. An
  organisation owner may rename it and manage its people; everybody else sees
  it and nothing more.
- A person who belongs to no organisation sees a screen that says so and
  offers to make one, rather than an empty list under a placeholder name.

## Done when

- Two organisations exist on the test instance and switching between them
  shows each one's own projects, with nothing leaking between them.
- Making an organisation from the screen works and lands on its projects.
- Renaming shows everywhere the name is shown, without a reload.
- An integration test proves somebody in one organisation cannot read the
  other's projects through the API, whatever the address says.
- No file under `web/src/` names an organisation.

## References

- `docs/BRIEFING-M01.md` section 3, the two audiences.
- `docs/design/reference/screenshots/desktop/projects.png`, the switcher.
- `internal/httpapi/router.go`, the organisation endpoints that exist.
- `docs/design/COMPARISON.md`, the projects screen row.

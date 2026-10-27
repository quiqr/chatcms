---
# chatcms-8n4m
title: 'Organisations: make, rename and switch, instead of one hardcoded slug'
status: completed
type: epic
priority: high
created_at: 2026-10-26T16:45:44Z
openspec-link: openspec/changes/archive/2026-10-27-add-organisations
updated_at: 2026-10-26T23:06:22Z
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-27-add-organisations`.

An organisation is now a thing the application knows about rather than a
constant it was compiled with. Every project route said `technative`, in
eight places, so a second organisation would have opened the first one's
work.

- **The address carries it.** A project is only unique within its
  organisation, so its address is `/o/:organisation/p/:project/...`. One
  request tells each screen who is looking, what they may switch to, and
  where they are; asking per screen would be several answers that can
  disagree.
- **An organisation screen**: its name, its address, and who is in it.
  Renaming keeps the address, because links already sent are written against
  it.
- **An organisation somebody is not in is nothing there**, not a redirect to
  one they are in. A redirect makes a wrong link look like a working one.
- **`check-interface` now fails if any source file names an organisation**, so
  this cannot come back quietly.

### Four defects this turned up

- **History called a change made minutes ago "Yesterday".** Each version is
  compared in the offset git recorded it in, against the server's own day.
  Caught only because the clock passed midnight during a run.
- **Refreshing the shell blanked every screen below it**, losing whatever
  somebody was in the middle of. The last answer is kept while a newer one is
  fetched.
- **An administrator could not see an organisation's people**, because
  reachability was decided before the permission was asked.
- **The projects screen asked an endpoint that lists only organisations you
  belong to**, so an administrator, who belongs to none, saw nothing.

### Still on the demo scaffold

The project shell still reads its project from `web/src/demo/content.ts`,
which is why the count on the Publish button can disagree with the publish
screen. That is the project settings epic, `chatcms-1r7q`.

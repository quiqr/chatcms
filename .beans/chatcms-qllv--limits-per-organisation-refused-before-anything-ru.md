---
# chatcms-qllv
title: Limits per organisation, refused before anything runs
status: completed
type: epic
priority: high
created_at: 2026-11-02T20:04:48Z
updated_at: 2026-11-02T23:30:00Z
parent: chatcms-7r0u
openspec-link: openspec/changes/archive/2026-11-02-add-organisation-limits
---

## Purpose

An organisation may have a maximum number of members and a maximum number of
projects. Both are optional, both are set by the platform admin, and both are
hard: anything that would take an organisation over a limit is refused before
it runs.

Without this the platform admin screens have numbers to show and nothing
behind them.

## Scope

- **Two optional limits per organisation**: members and projects. A new
  organisation has neither.
- **Refused before it runs**, server-side, for every guarded act: inviting
  somebody, adding a member, creating a project, importing one, duplicating
  one, and moving one in (counting the people who would become members of the
  destination).
- **The refusal says what happened, and the ways out.** The wording is in
  `docs/design/DESIGN.md` under "Limits (rules)": the limit, the numbers
  ("That makes 6 members, and the limit is 5"), and that somebody can remove
  something or ask the platform admin to raise or remove the limit. It
  carries a key, so the screen says it in Dutch too.
- **A pending invitation counts as a seat.** An archived project does not
  count as a project.
- **Lowering a limit below what is in use removes nothing.** It blocks what is
  new until usage is under the limit again.
- **Usage is shown the same way everywhere**: `n of max` with a 6 px bar,
  `primary` below 80%, `warning` from 80%, `destructive` at 100%, labelled
  "Almost full" and "Full". With no limit: `n · no limit`, and no bar.
- The primary button of a form that would be refused is disabled, with the
  reason in a `destructive-soft` box beside it, rather than a refusal only
  after pressing it.

## Done when

- Every guarded action is refused over its limit by the server, proven by an
  integration test per action.
- The same action is offered but disabled on the screen, with the reason,
  proven by an acceptance test for at least the two the design draws:
  duplicating at the project limit and inviting at the seat limit.
- A limit lowered below usage blocks additions and removes nothing.
- Usage reads the same on the organisation admin screens and the platform
  admin screens, because both ask the same function.

## References

- `docs/design/DESIGN.md`, "Limits (rules)" and sections 5.12 to 5.15.
- `docs/design/reference/html/org-projects-duplicate-at-limit.html`,
  `org-projects-move-at-limit.html`, `org-members-full.html`.

## Summary of Changes

An organisation may be allowed so many members and so many sites. Both are
optional, both belong to the platform admin, and both are refused before
anything runs.

- **Migration `00014`** adds the two nullable columns, with a constraint that
  a limit is at least one, and `projects.archived_at` so that the count the
  limit reads asks the right question from the start rather than being
  rewritten the day archiving lands.
- **`internal/tenancy/limits.go`** is the one place that reads how full an
  organisation is and the one place that refuses. The count and the act share
  a transaction, so two people taking the last seat cannot both get it; that
  needed `store.DB.Within(tx)`, which is new.
- **Three acts ask before they run**: putting somebody into an organisation,
  inviting somebody, and making a site. Changing the role of somebody already
  in a full organisation is deliberately allowed, because a full organisation
  must still be one somebody can reorganise.
- **A platform admin is not above the limit they set.** They are refused like
  anybody else; raising it is what they may do that nobody else can.
- **The refusal names the limit, the numbers and the ways out**, carries a
  key, and is therefore said in Dutch too.
- **`HowFull` and `whyThereIsNoRoom`** are one component, so the
  organisation's own screens and the platform admin's cannot disagree. The
  projects answer carries usage, because a New project button has to know
  whether there is room for one.
- Proved by six integration tests over the guard and the limits, two handler
  tests, five frontend tests, and
  `test/e2e/22-limits.spec.mjs` at 1440 px and at 390 px.

### What the gate caught

A stub in `internal/httpapi` stopped implementing an interface I had widened.
`go build ./...` does not compile tests, so only `go vet` inside the gate saw
it. Worth running vet after every interface change rather than after the last
one.

### Found on the way

A seat limit of one read as "All 1 seats are in use", and a mutating request
from a Playwright page is refused without the request-token header. The
first is plural-aware in both languages now; the second is a shared `asksFor`
helper in the fixtures rather than a trap each spec rediscovers.

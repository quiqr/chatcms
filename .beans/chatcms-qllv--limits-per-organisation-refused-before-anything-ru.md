---
# chatcms-qllv
title: Limits per organisation, refused before anything runs
status: in-progress
type: epic
priority: high
created_at: 2026-11-02T20:04:48Z
updated_at: 2026-11-02T21:48:11Z
parent: chatcms-7r0u
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

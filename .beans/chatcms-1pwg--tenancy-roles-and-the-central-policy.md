---
# chatcms-1pwg
title: Tenancy, roles and the central policy
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-22-add-tenancy-and-policy
updated_at: 2026-10-22T16:39:35Z
parent: chatcms-8vem
blocked_by:
    - chatcms-zijg
---

## Purpose

Organisations, projects and roles, with every authorisation decision going
through one function so per-document permissions can be added later without
hunting through handlers.

## Scope

- `internal/tenancy`: organisations, projects, memberships, invitations.
- `internal/policy`: the single `Can(actor, action, resource)` function. The
  roles are owner, implementer, editor and viewer, with the matrix from
  DESIGN.md section 6. The resource argument exists now even though milestone 1
  ignores most of it.
- Handlers call `Can`; nothing else compares roles. Add a test or lint that
  fails when a handler package references a role constant directly.
- Admin invites users by email with a role.
- The People and roles screen from DESIGN.md section 5.4, including the
  organisation and project segmented control, pending invites and resend.
- The organisation switcher and the Projects screen from DESIGN.md section 5.2.

## Done when

- A table test covers the full cross product of role, action and resource, with
  expected answers written out literally.
- A viewer cannot reach any mutating endpoint, proven by test, including by
  direct API call rather than only through the UI.
- Setup navigation and technical details are hidden, not disabled, for editors
  and viewers.
- Acceptance test 1 from the briefing passes at the API level.

## References

- `docs/BRIEFING-M01.md` section 4.2.
- `docs/design/DESIGN.md` sections 5.2, 5.4 and 6.

## Summary of Changes

Shipped as OpenSpec change `add-tenancy-and-policy`, archived at `openspec/changes/archive/2026-10-22-add-tenancy-and-policy`. Two capabilities built, `tenancy/organisations` and `tenancy/permissions`, with a requirement added to `platform/server-runtime`.

`internal/policy.Can` is pure: no database, no context, no error. That is what makes the whole table from DESIGN.md section 6 testable with the answers written out separately from the implementation, so a mistake in one does not become the expectation in the other. `TestOnlyPolicyNamesARole` walks the sources and fails when any other package carries a role as a quoted value; the allowed packages are listed explicitly. It caught the seed command during this change, which now uses the typed constants.

Owning an organisation carries the owner role on every project in it, so a project cannot end up with nobody who can administer it. A project in an organisation you are not in answers exactly as one that does not exist, so the API is not a directory of other customers sites.

Acceptance test 1 is proven end to end through the API: an organisation is created, a project made in it, a customer invited as an editor, and that customer sets a password through their link and signs in. A separate test calls every mutating endpoint as a viewer and checks that nothing changed.

The project list and the people and roles screen read real data. Checked in a browser at 1440 and 390 signed in: no sideways scrolling and no control under forty-four pixels.

Two decisions worth recording. The design foundation rendered both frames and let the stylesheet hide one; placed in both, a screen that loads mounts twice and asks the server twice, so one frame is now chosen from the same width the stylesheet uses. And the people table scrolled the whole page sideways on a phone: the cause was not the table but a screen-reader-only label inside it, absolutely positioned with no positioned ancestor, so its containing block was the page rather than the scrolling box.

Coverage: `internal/policy` 96%, `internal/tenancy` 79.9%, overall 85.8%. The gate no longer reports any core package as absent.

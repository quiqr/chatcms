---
# chatcms-1pwg
title: Tenancy, roles and the central policy
status: todo
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:31Z
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

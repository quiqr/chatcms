---
# chatcms-rbdb
title: An application administrator, who creates and manages organisations
status: todo
type: epic
priority: critical
created_at: 2026-10-26T21:01:46Z
updated_at: 2026-10-26T21:01:46Z
parent: chatcms-8vem
---

## Purpose

ChatCMS is hosted and multi-tenant, and nothing in it is above an
organisation. `Service.CreateOrganisation` never asks `policy.Can`, so any
signed-in person can make as many organisations as they like and owns every
one of them. There is no public sign-up, so today that means any invited
person rather than anybody at all, but an editor invited to fix a typo on one
bakery site should not be able to do it.

Acceptance test 1 in the briefing begins "As admin". This epic is the admin.

## Scope

- An application administrator, held on the person, not on a membership. A
  migration adds it; `policy.Actor` carries it; `policy.Can` is still the one
  place that decides.
- Creating an organisation needs it. Renaming and removing one needs it.
  Nobody else may, and the refusal comes from the policy rather than from a
  hidden button.
- An administrator sees every organisation, who is in each, and can put
  somebody in or take them out. That is what "manages organizations" means
  here.
- **An administrator is not an editor of everybody's site.** Managing who may
  reach a site is not the same as reaching it, and a hosted CMS whose staff
  silently hold edit rights on every customer's content is one no customer
  should accept. An administrator who needs to read a customer's content
  joins that organisation, which is visible to its owner. The alternative,
  letting the flag pass every check, is simpler and worse, and this epic
  argues that in its proposal rather than deciding it quietly.
- Appointing the first one, which cannot itself need an administrator: the
  server takes an email address from its configuration, as a value or from a
  file like every other secret, and makes that account an administrator on
  start-up. Idempotent, so it is safe in a NixOS module that runs on every
  boot, and it says in the log what it did.
- Administrators appoint and remove other administrators. The last one cannot
  remove themselves, because an installation with no administrator can only be
  repaired with SQL.
- Every administrator action is recorded with who did it and when. Somebody
  with power over all tenants leaves a trail.
- The seed makes one, so development and the acceptance tests have an
  administrator to be.

## Done when

- Somebody who is not an administrator is refused when creating an
  organisation, by `policy.Can`, proven in both a unit test and an integration
  test against the real endpoint.
- An administrator creates an organisation, puts an owner in it, and that
  owner sees it; the administrator does not appear in its people unless put
  there.
- An administrator cannot read or change a project's content without joining
  the organisation, and a test says so.
- A fresh database plus the configured address yields exactly one
  administrator, and starting twice does not change that.
- Removing the last administrator is refused.
- Acceptance test 1 runs as an administrator rather than as a seeded owner.

## References

- `docs/BRIEFING-M01.md` section 6, acceptance test 1.
- `internal/policy/policy.go`, `Can` and `EffectiveRole`.
- `internal/tenancy/service.go:57`, which creates an organisation unchecked.

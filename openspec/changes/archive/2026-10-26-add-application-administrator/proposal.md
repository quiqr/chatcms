## Why

ChatCMS is hosted and multi-tenant, and nothing in it is above an
organisation. `internal/tenancy/service.go` creates one without asking
`policy.Can` at all, so any signed-in person can make as many as they like and
owns every one. There is no public sign-up, so today that means any invited
person rather than anybody at all, but an editor invited to fix a typo on one
bakery site should not be able to do it.

The briefing's first acceptance test begins "As admin". This change is the
admin. Epic `chatcms-rbdb`.

Bean: `.beans/chatcms-rbdb--an-application-administrator-who-creates-and-manag.md`.

## What Changes

- **An administrator is a property of a person**, not a membership, because it
  is held over the whole installation rather than over one organisation.
- **Making, renaming and removing an organisation needs it.** The refusal comes
  from `policy.Can`, which stays the only place a permission is decided.
- **An administrator is not an editor of every site.** They see every
  organisation and who is in it, and may put people in and take them out.
  They do not thereby gain the right to read or change a customer's content.
- **The first administrator is appointed by configuration**, as a value or from
  a file, applied on start-up, idempotently.
- **Administrators appoint each other**, and the last one cannot be removed.
- **What an administrator does is recorded**, with who and when.

## Capabilities

### New Capabilities

- `tenancy/administration`: who may hold the installation, what it lets them
  do, what it deliberately does not, and how the first one comes to exist.

### Modified Capabilities

- `tenancy/organisations`: creating one now names who may.
- `tenancy/permissions`: the decision function gains a level above the
  organisation.

## Impact

- **New**: `internal/admin`, a migration adding the flag and the record of
  what was done, endpoints under `/api/administration`, and the screens for
  them.
- **Changed**: `internal/policy` carries the flag and answers for it;
  `internal/tenancy` asks before creating; `cmd/seed` makes one; the NixOS
  module and `scripts/dev.sh` pass the first administrator's address.

## Assumptions

- **Managing a tenant is not entering it.** The simple implementation is a
  flag that passes every check in `policy.Can`. It is one line and it is
  wrong: it would silently give whoever runs chatcms.eu edit rights on every
  customer's site, and a customer cannot tell that it is so. An administrator
  who must reach content joins that organisation, which its owner can see.
  The operator can of course read the database directly; the difference is
  that doing it through the product would look like ordinary use and leave no
  trace that the customer could find.

- **The flag lives on the person, not on a role table.** It is not a role on a
  resource; it is a property of an account. Modelling it as a membership of a
  pretend organisation would make `EffectiveRole` lie about what somebody
  holds.

- **Appointing the first one cannot need one.** A new installation has no
  administrator to appoint anybody, so the server reads an address from its
  configuration and promotes that account when it starts. It is idempotent,
  because a NixOS module runs it on every boot, and it says in the log
  whether it changed anything.

- **The address must already be a person.** Promoting on start-up does not
  create an account, because an account created without a password or an
  invitation is an account nobody can use and a name nobody chose. If there is
  no such person yet, the server says so and carries on; the next start, after
  an invitation has been accepted, promotes them.

- **The last administrator cannot be removed, including by themselves.** An
  installation with none can only be repaired with SQL on the database, which
  is not a thing a customer can be asked to do.

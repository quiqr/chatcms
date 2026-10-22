## Why

Knowing who someone is was the last change; knowing what they may do is this
one. Until it lands, every signed-in person can reach everything, and there is
nothing for a project to belong to. Epic `chatcms-1pwg`.

It completes acceptance test 1, "as admin, create an organisation and invite
the customer as editor", and it is the precondition for onboarding, which is
the precondition for publishing.

## What Changes

- **There are organisations, and projects inside them.** A person belongs to
  an organisation, and holds a role on each project they can reach.
- **Four roles, with the powers the designs give them.** Owner, implementer,
  editor and viewer, exactly as `docs/design/DESIGN.md` section 6 sets out.
- **One function decides everything.** `Can(actor, action, resource)` is the
  only place in the codebase where a role is compared. A test over the sources
  fails the build when any other package mentions one.
- **The decision takes a resource it mostly ignores.** Per-document
  permissions are a later milestone, and the argument exists now so that
  adding them is a change to one function rather than to every caller.
- **An owner invites someone by address and role.** The invitation creates the
  person if they are new, gives them the role, and issues the link that sets
  their first password.
- **A viewer cannot change anything**, and this is proven by calling the API
  directly rather than by hiding a button.
- **What a role cannot use is absent, not greyed out.** The interface is told
  what the person may do, and leaves out what they may not.
- **Two screens exist**: the project list with its organisation switcher, and
  people and roles.

## Capabilities

### New Capabilities

- `tenancy/organisations`: what an organisation and a project are, who belongs
  to them, and how someone is invited.
- `tenancy/permissions`: who may do what, and the single place that decides.

### Modified Capabilities

- `platform/server-runtime`: endpoints now refuse a request whose actor may
  not make it, with an answer that does not reveal what exists.

## Impact

- **New code**: `internal/policy`, `internal/tenancy`, their handlers, and the
  two screens.
- **Changed code**: `internal/httpapi` gains the middleware that turns a
  refusal into an answer; `internal/auth` gains the invitation path that
  creates a person.
- **Schema**: one migration adding organisations, projects, and the two kinds
  of membership.
- **Dependencies added**: none.
- **Seams**: this change implements the third seam named in
  `docs/ARCHITECTURE.md`, and is the last of the four.
- **Coverage**: `internal/policy` becomes the last core package to carry the
  80% floor, so the gate stops reporting any as absent.

## Assumptions

- **An organisation owner owns every project in it.** Rather than giving a
  person a row on each project, membership of the organisation as an owner
  grants the owner role throughout. It is one rule, it matches how a small
  agency actually works, and it means a project cannot end up with nobody who
  can administer it.
- **A project's own members are the four roles.** Organisation membership is
  owner or member; a member reaches only the projects they hold a role on.
- **Inviting creates the person.** An address that belongs to nobody becomes a
  person with no password and an invitation link. An address that is already
  known is given the role and told, rather than being made a second time.
- **Nothing is deleted yet.** Removing a person from a project is in scope;
  deleting a project or an organisation is not, because it destroys content
  and belongs with a confirmation flow the designs put in project settings.
- **Billing is named in the role table and does not exist.** The designs give
  owners billing; there is nothing to bill. The action is defined so the table
  is complete and the screens can hide what they must.
- **Two parts of the people screen wait for what records them.** The designs
  show a "This project / Organisation" switch and a "Last active" column.
  Organisation-level membership management is not in this change, and nothing
  records when anyone was last active. Both are left out rather than filled
  with something invented.
- **A decision from the design foundation is revised here.** That change
  rendered both frames and let the stylesheet hide one, which is right for
  chrome and wrong for a screen that loads: placed in both frames, a screen
  mounts twice and asks the server twice. One frame is now chosen, read
  synchronously from the same width the stylesheet uses, so there is still
  nothing to flash. The design document records it.

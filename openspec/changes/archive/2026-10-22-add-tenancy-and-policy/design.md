## Context

`internal/auth` puts a person on every request. `internal/httpapi` hands that
person to handlers. What is missing is anything for the person to have a role
on, and anywhere for the question "may they?" to be answered. See
`proposal.md`.

`docs/design/DESIGN.md` section 6 is the specification of the role table, and
`docs/ARCHITECTURE.md` names the seam: one `Can` function, handlers that ask
rather than decide.

## Goals / Non-Goals

**Goals:**

- A permission can be changed in one place and cannot be forgotten in another.
- A viewer cannot change anything, provable by calling the API rather than by
  looking at the interface.
- Per-document permissions can be added later by changing one function.

**Non-Goals:**

- Deleting projects or organisations. See the assumption in `proposal.md`.
- Billing. The action exists; nothing bills.
- Sending the invitation. The mail epic owns delivery; the link goes to the
  log as it does for every other link today.

## Decisions

### `Can` is pure, and takes everything it needs

```go
func Can(actor Actor, action Action, resource Resource) bool
```

No database, no context, no error. The `Actor` carries the person's
memberships, which the caller has already loaded; `Resource` names the
organisation and project concerned and, one day, something inside it.

A pure function is what makes the full table testable: every role against
every action, with the expected answer written out rather than computed. A
function that reached for the database could not be tested that way, and the
table is the specification.

The cost is that the caller has to load memberships first. That is one query,
cached per request, and it is the right place for it: loading is I/O and
deciding is not.

### The table is data, and the test reads it the other way round

The actions are a list; the matrix is a map from role to the actions it
allows. The test asserts every role against every action using an expected
table written separately, so that an error in the implementation's table does
not silently become the test's expectation too.

### Nothing outside `internal/policy` may name a role

Enforced the way the git rule is: a test walks the sources and fails when a
package outside `internal/policy` mentions a role name. The exceptions are the
tenancy package, which stores roles, and the generated query code, which
carries the column.

Those exceptions are listed explicitly rather than pattern-matched, so adding
one is a visible change.

### Owning an organisation owns its projects

A person's effective role on a project is the highest of: the role they hold
on the project, and owner if they own the organisation. One rule, and it means
a project cannot end up with nobody who can administer it, which is the
failure mode that generates support requests.

### Refusals do not reveal what exists

A project in an organisation the person is not in answers the same as a
project that does not exist. Otherwise the API is a directory of every
customer's project names.

For a project they can see but an action they may not take, 403 is right and
says so: hiding that would make the interface unable to explain itself.

### Invitations reuse the credential links already built

`internal/auth` already issues single-use expiring links and records them in
the log. Inviting someone creates the person with no password, gives the role,
and issues an invitation link through that same path. No second mechanism.

### The interface is told, not left to work it out

A project's own view carries a list of what this person may do. The frontend
hides what is absent rather than re-deriving the role table in TypeScript,
which would be the role table existing twice and drifting.

### One frame is chosen, not both rendered

The design foundation decided to render the project frame and the phone frame
together and let the stylesheet hide one. That is right for chrome and wrong
for anything that loads: the first screen placed in both frames mounted twice
and asked the server twice.

A frame is now chosen with `matchMedia`, read synchronously on first render so
there is nothing to flash, against the same width the stylesheet uses, named
in one place so the two cannot drift.

### Sideways scrolling, and where it actually came from

The people table is wider than a phone and scrolls inside its own box, as the
hand-off asks. The page nonetheless scrolled sideways, and the cause was not
the table: a screen-reader-only label inside it is absolutely positioned, and
with no positioned ancestor its containing block was the page rather than the
scrolling box, so it sat half a screen to the right and dragged the page with
it. The scrolling box is now positioned, which contains it.

Worth writing down because the symptom pointed at the table and the cause was
a utility class three levels down.

## Risks / Trade-offs

- **The memberships load on every request** that asks a permission. One query,
  and it is cached for the life of the request. If it ever matters, it is a
  cache, not a redesign.
- **The source-walking rule is a heuristic**, as the git one is. A role named
  through a variable escapes it. The cost of the escape is low because the
  only reason to name a role elsewhere is to compare it, which is what review
  then catches.
- **Organisation ownership is coarse.** An organisation owner can reach every
  project in it, including ones they were never added to. That is the intent,
  and it is stated in the proposal so it is a decision rather than a surprise.
- **Nothing can be deleted**, so a mistake leaves a project lying around.
  Acceptable for a milestone that is about getting a real customer running;
  deletion needs the confirmation flow the designs put in project settings.

## Migration Plan

One migration adds organisations, projects and the two memberships. Nothing is
deployed. The first organisation and its owner are created by the deployment
epic's command; until then, tests create them directly.

## Open Questions

- Whether an organisation needs roles beyond owner and member. Milestone 1
  does not need them, and adding one is a value in a column.

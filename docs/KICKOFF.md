# Kickoff

What to hand Claude Code so it can build this autonomously, and what to check
while it does.

## Starting a session

```
Read AGENTS.md, docs/BRIEFING-M01.md, docs/ARCHITECTURE.md and
docs/TESTING.md. Then run `beans list --ready` and take the next epic.

For that epic: write an OpenSpec change proposal that links the bean id, set
the bean to in-progress, then run /mip:ship for the change. Do not stop between
steps and do not ask for confirmation on anything the briefing, the design or
AGENTS.md already decides. Log assumptions in the proposal.

When the change is shipped, close the bean with a Summary of Changes and start
the next ready epic.
```

That prompt is repeatable. Each run takes one epic from proposal to pushed
commit, so a session can be stopped and resumed at any epic boundary.

## The loop, once per epic

1. `beans list --ready` picks the epic. Dependencies are already wired, so the
   order takes care of itself.
2. `/openspec-propose` writes `openspec/changes/<name>/` with the proposal, the
   spec deltas and `tasks.md`. The proposal links the bean; the bean goes to
   in-progress.
3. `/mip:ship <name>` implements every task with tests, updates the CHANGELOG,
   runs `scripts/ship-change.sh`, which gates on `nix flake check` before it
   archives, commits and pushes, and then closes the bean.
4. The bean gets `openspec-link: openspec/changes/archive/...` in its
   frontmatter.

Nothing archives, commits or pushes while the gate is red, so a failed run
leaves the tree inspectable rather than half-shipped.

## What to check between epics

| Check                            | Command                             |
| -------------------------------- | ----------------------------------- |
| The gate is green                | `nix flake check`                   |
| Coverage did not quietly slip    | the gate prints per-package numbers |
| The backlog moved                | `beans roadmap`                     |
| Specs match what was built       | `openspec list`, then read the spec |
| Nothing was committed by Claude  | `jj log` shows Pim Snel only        |

## Where it will want to ask you something

These are the open questions the briefing does not settle. Answering them
before the matching epic saves a round trip.

- **The customer site.** Acceptance tests run against a real customer repository
  with a Quiqr model. Until that URL and a deploy key exist, the end-to-end
  epic runs against fixtures only.
- **Email delivery.** Invitations and password resets need an SMTP target. Until
  one exists, the auth epic will log the link instead of sending it, which the
  proposal should record as an assumption.
- **The golden source.** Goldens are generated from the old Quiqr desktop app at
  `../quiqr-desktop`. If that checkout is not the version the customer actually
  used, the goldens pin the wrong behaviour.
- **Hosting.** The NixOS module assumes a host with Postgres and a reverse proxy
  in front. The deployment epic needs to know which host, and where the secret
  key file comes from.

## The first ship will be the slowest

The foundation epic adds the frontend to the flake, which means a pinned
`npmDepsHash` and the first full Nix build of the Node dependency tree. After
that the gate is fast enough to run on every change.

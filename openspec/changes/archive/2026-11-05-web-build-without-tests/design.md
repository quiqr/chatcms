# Design

## Context

See proposal.md. `mkWeb` had `doCheck = true`, and `mkServer` copies
`mkWeb args` into `internal/webui/dist`, so the check and the shipped build
were one derivation.

## Decisions

### A flag on `mkWeb`, not a second derivation

`checked ? false` sets `doCheck`. The shipped build and the check share the
source and the npm dependencies, and differ only in whether the check phase
runs. The alternative, raising Vitest's timeout, would make the gate slower to
notice a hanging test and still run the suite on every deploy.

## Risks / Trade-offs

- [A frontend that fails its checks is built and embedded] → Only outside the
  gate: `ship-change.sh` runs `nix flake check` before anything is committed
  or pushed, and mipnix deploys a pushed commit.

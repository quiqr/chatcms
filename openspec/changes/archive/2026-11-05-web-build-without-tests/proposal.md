# Proposal

## Why

The first deploy to durer through mipnix failed on a Vitest test that timed
out at five seconds. The server embeds the same frontend derivation that the
`frontend` check builds, so every deployment ran the whole frontend suite again
on the deploying machine, under the load of building a full NixOS system. A
deploy should not fail on a test the gate already passed.

Originating bean: none. Found while deploying to durer, as with chatcms-xkg0.

Acceptance test served: none directly; it unblocks deploying the MVP that
section 6 of `docs/BRIEFING-M01.md` is tested against.

## What Changes

- The frontend build that `chatcmsd` embeds, and `packages.web`, no longer run
  the type check, the interface and string checks, the bundle check or Vitest.
- `checks.frontend` still runs all of them, so `nix flake check` and
  `scripts/ship-change.sh` gate on exactly what they gated on before.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. What the gate enforces does not change; only where it runs does.

## Impact

- `flake.nix`: `mkWeb` takes a `checked` flag, off by default and on for
  `checks.frontend`.
- Deployments through mipnix build the frontend without its tests.

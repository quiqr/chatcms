## Why

The change `add-quiqr-model-parser` promised three capabilities and shipped
two. `content/model-schema` was listed in its proposal, built, and covered by
tests, but its spec file never reached the change directory, so the archive had
nothing to apply. The behaviour is in `internal/model`; only the written
contract is missing.

Milestone 01's definition of done includes "all specs in OpenSpec match what
was built", so this is a gap that has to close now rather than at the end.
Epic `chatcms-vho6`.

## What Changes

Nothing in the code. This change adds the `content/model-schema` capability
spec describing what is already implemented: what an assembled model contains,
which field types are recognised, how an unrecognised type and an unrecognised
key survive, and what an invalid model reports.

Each requirement below is already proven by a test in
`internal/model`, named in the task list, so the spec documents behaviour
rather than proposing it.

## Capabilities

### New Capabilities

- `content/model-schema`: what an assembled model contains once typed, which
  field types exist, and what is preserved or rejected.

### Modified Capabilities

None.

## Impact

- **Code**: none.
- **Specs**: one new capability under `openspec/specs/content/model-schema/`.
- **Process**: `scripts/ship-change.sh` gates on `nix flake check`, which does
  not notice a spec file that was never written. The task list adds a check
  that every capability a proposal declares has a matching spec file before the
  change can ship.

## Assumptions

- **The spec is written from the tests, not from memory.** Every scenario is
  matched to an existing test, and anything the tests do not cover is left out
  rather than asserted.

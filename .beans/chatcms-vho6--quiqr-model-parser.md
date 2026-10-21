---
# chatcms-vho6
title: Quiqr model parser
status: completed
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-21-add-quiqr-model-parser
updated_at: 2026-10-21T17:28:43Z
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

Read the Quiqr model. This is the heart of the product: the declarative YAML
that decides which forms exist and what they contain. It must read today's
models unchanged.

## Scope

- `internal/model`: parse the Quiqr model YAML from a site's quiqr directory
  into a typed schema: singles, collections, field groups and every existing
  field type.
- Derive the content navigation from the model: Pages and Collections with
  counts, and the item name used in labels such as "New article".
- Serve the schema to the frontend as JSON, so form rendering is data driven.
- Validation with errors that name the file, the path in the model and what was
  expected.
- An inventory document listing every field type the old Quiqr supports, with
  its semantics, written from the reference app at `../quiqr-desktop`.

## Done when

- Every model in `../quiqr-desktop` test resources and in the customer site
  parses without error.
- Unknown keys are preserved rather than dropped, so round-tripping a model
  does not lose information.
- `internal/model` is at or above 80% statement coverage.
- The field type inventory is committed and referenced by the form epic.

## References

- `docs/BRIEFING-M01.md` sections 2 and 4.4.
- The reference app at `../quiqr-desktop`.

## Summary of Changes

Shipped as two OpenSpec changes: `add-quiqr-model-parser` (archived at `openspec/changes/archive/2026-10-21-add-quiqr-model-parser`) and `add-model-schema-spec` (archived at `openspec/changes/archive/2026-10-21-add-model-schema-spec`).

Four capabilities are specified and built: `content/model-source`, `content/model-schema`, `content/content-navigation`, and the spec-completeness check that closed a gap found while shipping.

`internal/model` now reads a site's model the way the desktop app does: base file in YAML, TOML or JSON; includes merged by filename with the base winning; per-item include folders; and partials with the item winning and the reference application's exact field deduplication. It is a pure read through an `fs.FS`, so the git working copy can be handed to it later unchanged.

All twenty-seven field types are typed, each with its own configuration struct where it has properties of its own. An unrecognised type survives as a placeholder naming what it declared; unrecognised keys survive on the model, on singles, on collections and on fields. Errors carry file, path and expectation.

Content navigation is derived from the model: pages and collections, item counts read from the working copy, the model's own singular item names, the declared menu order, hidden entries omitted but still reachable by key.

`docs/QUIQR-MODEL.md` is the written inventory, sourced line by line from the reference app. Six fixtures under `test/fixtures/models/` pin the JSON shape with golden files.

The desktop app's bundled `dog_food_model` is deliberately excluded and reported as a warning; remote partials are read from cache rather than fetched. Both are recorded as assumptions in the proposal.

Coverage: `internal/model` at 91.9%, overall 89.8%.

One thing went wrong worth recording: the first change shipped with two of its three promised capability specs, because a spec file was never written and nothing noticed. `scripts/check-specs.sh` now runs before the gate in `scripts/ship-change.sh` and fails when a proposal declares a capability it has no spec for.

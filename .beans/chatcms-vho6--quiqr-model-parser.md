---
# chatcms-vho6
title: Quiqr model parser
status: todo
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:31Z
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

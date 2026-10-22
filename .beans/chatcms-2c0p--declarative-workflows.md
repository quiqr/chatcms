---
# chatcms-2c0p
title: Declarative workflows
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T19:24:34Z
parent: chatcms-8vem
blocked_by:
    - chatcms-dupp
---

## Purpose

Declarative workflows so a site can describe what happens on save, preview and
publish without the app hard-coding one pipeline.

## Scope

- Workflow files as YAML in a `workflows` subdirectory of the site's quiqr
  directory.
- `internal/workflows`: parse, validate and run them. Triggers are save,
  preview and publish. Steps call built-in actions.
- Built-in actions for milestone 1: hugo build, pre-publish, git push.
- The pre-publish step is optional and may be empty. When it fails nothing is
  pushed and the user gets a friendly error naming what to do.
- Steps run through `internal/runner`, so a workflow cannot escape the sandbox.
- Validation errors name the file and the step.

## Done when

- A site with no workflow file publishes using sensible defaults.
- A failing pre-publish step blocks the push and produces the error state from
  DESIGN.md section 5.8, with the raw log only for owners and implementers.
- Workflow parsing and execution are covered by tests including the failure
  paths.

## References

- `docs/BRIEFING-M01.md` section 4.10.
- `docs/design/DESIGN.md` section 5.8.

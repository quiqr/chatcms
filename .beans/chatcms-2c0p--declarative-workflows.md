---
# chatcms-2c0p
title: Declarative workflows
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T23:10:00Z
openspec-link: openspec/changes/archive/2026-10-22-add-declarative-workflows
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-22-add-declarative-workflows`.

A site can now say what happens on save, preview and publish, in YAML under
`quiqr/workflows/`:

```yaml
on: publish
steps:
  - name: Checking pages and links
    action: check
    run: ./scripts/check-links.sh
  - action: build
  - action: publish
```

- A key the format does not define is a mistake rather than something to
  ignore, and every error names the file, the step and the key. A workflow is
  written by hand once and then forgotten, which is exactly when a silently
  ignored `step:` instead of `steps:` costs an afternoon.
- Publishing has to be last and can only happen once; building can only
  happen once. A workflow that pushes and then builds has a typo or a
  misunderstanding in it, and neither is served by running it.
- A site that declares nothing still publishes. The default is worked out
  from the project's own publishing settings: a project whose pipeline builds
  the site is not built here, and one that publishes built pages builds
  first. A site that declares a workflow with no steps gets that, not the
  default, because saying nothing happens is a thing a site may say.
- `internal/workflows` runs the steps in order and stops at the first
  failure. What ran before is still reported.
- A check runs in the sandbox against the site's files, with a fixed pinned
  toolbox, no network, a time limit and a memory limit. A test runs a real
  check that writes into the site and shows that nothing it wrote reaches
  what is published.
- A failure carries two things kept apart in the type: a sentence for the
  person who pressed the button, with no git word in it, and the raw output
  on the step. A test walks every way a step can fail and checks the sentence
  against the forbidden-word list from `docs/design/DESIGN.md` section 7.

Publishing is not a sandboxed step, and the architecture doc says so in a
table rather than leaving "every step runs in the sandbox" to be assumed. It
needs the network and the deploy key, which is everything the sandbox exists
to withhold, and it runs no code the site wrote.

What lands on the publishing branch is left to the publish epic, which this
change gives a seam to implement: a publisher is handed the project and, when
a build ran, the pages it produced. The server wires a publisher that refuses
rather than pretending, so no path can report a publish that never happened.

41 tests, no skips inside the Nix build sandbox. Coverage 95.3% on the
package, 86.1% overall.

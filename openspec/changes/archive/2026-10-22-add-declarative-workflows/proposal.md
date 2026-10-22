## Why

A site needs a say in what happens when it is published. One site runs a link
check first; another builds with Hugo; another pushes source and lets its own
pipeline build. Hard-coding one pipeline would make ChatCMS wrong for the
second customer. Epic `chatcms-2c0p`.

It is also the last thing publishing waits on, which is acceptance test 7.

## What Changes

- **A site can declare what happens**, in YAML files under
  `quiqr/workflows/`, with one workflow per trigger: save, preview, publish.
- **A site that declares nothing still works.** Every trigger has a default
  worked out from the project's own publishing settings, so an imported site
  publishes correctly before anybody writes a workflow.
- **Three built-in actions**: build the site with Hugo, run the site's own
  check before publishing, and publish.
- **The site's own check runs in the sandbox**, with a fixed toolbox, no
  network and the same limits as a build. It is the only step that runs code
  the site wrote.
- **A failed check publishes nothing.** The workflow stops, the site is
  untouched, and what comes back is a sentence an editor can act on, with the
  raw log kept separately for people allowed to see technical detail.
- **A mistake in a workflow file names the file and the step.** Unknown keys
  are mistakes rather than ignored.

## Capabilities

### New Capabilities

- `workflows/definition`: what a workflow file may say, what it means, and
  what happens when it says something wrong.
- `workflows/running`: what happens when a trigger fires, in what order, and
  what comes back.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/workflows`.
- **Schema**: none. A workflow lives in the site, not the database; a run of
  one is recorded through the existing build record.
- **Settings**: the toolbox a check may use, and the limits a check runs
  under.
- **Depends on**: `internal/runner` and `internal/sitebuild` from
  `add-sandboxed-runner`.

## Assumptions

- **The toolbox is fixed and pinned.** A check gets a shell and the usual
  utilities, named by a setting and supplied by Nix. A site cannot ask for
  another tool in milestone 1. Letting a workflow name its own dependencies
  is the shareable-actions idea the briefing puts out of scope, and doing it
  badly now means fetching software from the internet during a publish.

- **Publishing is not a sandboxed step.** It needs the network and the
  project's deploy key, which is everything the sandbox exists to withhold.
  It runs no code the site wrote, so there is nothing to contain: it is
  ChatCMS doing what the project settings say. The sandbox bounds the site's
  own code, which is the check and the build.

- **Pushing built HTML is left to the publish epic.** This change gives the
  publishing step a seam to call and the default workflow that builds first.
  What exactly lands on the publishing branch is the publish pipeline's
  business, and it is already an epic of its own.

- **One workflow per trigger.** Two files claiming the same trigger is a
  mistake worth naming rather than an ordering puzzle to solve. The error
  names both files.

- **Save and preview do nothing by default.** Saving already makes a version
  and preview already renders; neither needs a step. The triggers exist so a
  site can add one, and so the vocabulary does not have to change later.

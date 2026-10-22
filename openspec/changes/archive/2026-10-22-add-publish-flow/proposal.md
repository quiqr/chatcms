## Why

Publishing is the one deliberate act in the product. Everything else saves
quietly; this is the moment somebody decides the world should see it. It has
to show what is about to happen in words a person recognises, and it has to
be safe to press. Epic `chatcms-5uu2`, acceptance test 7.

## What Changes

- **A review before anything happens**: what changed since the last publish,
  as pages and fields rather than as a list of files, with the words that
  changed shown struck through and highlighted.
- **An optional note**, "What's in this update?", kept with the publish so
  History can show it later.
- **Publishing** sends either the site's own files or the pages ChatCMS
  built, to the main line or to a line of its own, as the project's settings
  say.
- **Built pages go to a line of their own** without making a version of the
  site's source and without touching the working copy, so somebody can keep
  editing while it happens.
- **A failed publish leaves the live site exactly as it was** and says so in
  words with no git in them, with the raw output kept apart for people
  allowed technical detail.
- **Every publish is recorded**: who, when, what it carried, how it went.
- **The screen** for all four states the designs draw, working at 390 px.

## Capabilities

### New Capabilities

- `publish/review`: what somebody is shown before deciding, and in what words.
- `publish/act`: what publishing does, what is recorded, and what happens
  when it fails.

### Modified Capabilities

- `sync/repository`: the git engine learns to say what differs from the
  published site, to read a file as the live site has it, and to publish to a
  named line, in source or as built pages.

## Impact

- **New code**: `internal/publish`, `web/src/routes/PublishPage.tsx`.
- **Changed**: `internal/gitsync` gains four methods; `internal/sitebuild`
  hands back the built site rather than the folder holding it.
- **Schema**: one migration adding the publishes table.
- **Depends on**: `add-declarative-workflows`, which is what a publish runs.

## Assumptions

- **A publish is one request, start to finish.** The designs draw four steps
  with a progress bar, which suggests a publish somebody watches. Milestone 1
  publishes a small site in a few seconds, so the request carries it and the
  answer carries every step with its own outcome. Making it a background job
  with its own progress stream is work the queue already supports and nothing
  yet needs. The record is written at the start and finished at the end, so
  the move costs no schema change.

- **Only one publish per project at a time.** The database refuses a second
  while one is running, rather than this being watched for in code, because
  two publishes racing to the same line is the one failure that could leave
  the live site in a state nobody asked for.

- **A changed file the model says nothing about is still listed.** It is
  named by its path and badged as a change. A change nobody is shown is worse
  than one shown plainly, and an implementer editing a layout should see that
  it is about to go out.

- **The word diff compares words, not characters.** A person wants to see
  which words changed. Very long text is marked as replaced wholesale rather
  than compared, because the comparison is quadratic and a page body is not
  worth a second of somebody's time.

- **Publishing built pages over the site's own line is refused.** The two
  settings allow the combination and it would replace a customer's source
  with its own output. The git engine refuses it and says why.

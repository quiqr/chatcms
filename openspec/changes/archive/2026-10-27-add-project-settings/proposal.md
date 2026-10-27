## Why

`/o/:organisation/p/:project/settings` is a heading with nothing under it.
Everything decided when a project was set up can only be decided once: the
setup screen is reachable only while a project has no site. An implementer
who picked the wrong line to publish to, or a customer who moved host, has no
way to change it.

The shell around every project screen also still reads its project from
`web/src/demo/content.ts`, which is why the count on the Publish button can
disagree with the publish screen beside it.

Epic `chatcms-1r7q`.

Bean: `.beans/chatcms-1r7q--project-settings-change-after-setup-what-setup-dec.md`.

## What Changes

- **A project settings screen**: its name, the address shown to customers,
  where the site lives, and how it is published, all changeable afterwards.
- **The panels are the setup screen's**, reused rather than copied, so the two
  cannot drift.
- **Changing how it is published says what the next publish will do
  differently** before it is saved.
- **The shell reads the project from the server**, so the count on the Publish
  button is the real one.
- **Removing a project**, with a confirmation naming what is removed and what
  is not: the site in git is not touched.
- **Technical detail stays behind the same permission** it is behind
  everywhere else.

## Capabilities

### New Capabilities

None. This gives the projects capability the screen its rules already
describe.

### Modified Capabilities

- `tenancy/projects`: a project's name, address, source and publishing can be
  changed after it is set up, and a project can be removed.

## Impact

- **New**: `web/src/routes/ProjectSettingsPage.tsx`, endpoints for changing a
  project and for removing one.
- **Changed**: `internal/tenancy` and `internal/onboarding` gain the changes;
  `ProjectShell`, `PhoneShell` and `EditorShell` read the project from the
  server rather than from the demo scaffold.
- **Removed**: the demo project scaffold, once nothing reads it.

## Assumptions

- **The address a project is reached at does not change.** Renaming is safe;
  changing the slug would break every link already sent, and milestone 01 has
  no way to keep the old one working. The screen says the address is fixed
  rather than offering a change that quietly breaks things.

- **Changing where the site lives is not re-importing it.** The work already
  brought in stays; only where it is sent changes. A project whose address in
  git is changed to an unrelated repository is somebody's mistake, so the
  screen says what will happen and tests the connection first.

- **Removing a project leaves the site alone.** What is in git is the
  customer's, and ChatCMS forgetting a project must never be a way to lose
  it. The confirmation says so in as many words.

- **The count on the Publish button comes from the same answer the publish
  screen uses.** Two sources for one number is how they came to disagree.

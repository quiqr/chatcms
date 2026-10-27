## Why

An organisation can be created over the API and listed, but no screen makes
one, renames one, or shows who is in it. Worse, the frontend does not know
which organisation it is looking at: every project route in `web/src/App.tsx`
passes `organisation="technative"`, which the seed happens to create, in
eight places. A second organisation would therefore open the first one's
data, and the shell around every screen reads its name, its projects and who
you are from `web/src/demo/content.ts` rather than from the server.

Epic `chatcms-8n4m`.

Bean: `.beans/chatcms-8n4m--organisations-make-rename-and-switch-instead-of-on.md`.

## What Changes

- **The organisation comes from the address.** Routes carry it, the shells
  read it, and no file under `web/src/` names one.
- **The shell shows the real thing**: the organisation you are in, the
  organisations you may switch to, and your own name, all from the server.
- **An organisation screen**: its name, its address, the people in it with
  their organisation role, and renaming.
- **Renaming** needs administering, as `chatcms-rbdb` settled. Managing the
  people in one is the owner's.
- **Somebody in no organisation** is told so, and told what to do about it,
  rather than shown an empty list under a placeholder name.

## Capabilities

### New Capabilities

None. This change gives the organisations capability its screens and makes
the existing rules reachable.

### Modified Capabilities

- `tenancy/organisations`: renaming, and who may see and manage the people in
  one.

## Impact

- **New**: `web/src/lib/whoami.ts` and the context the shells read,
  `web/src/routes/OrganisationSettingsPage.tsx`, endpoints for renaming and
  for an organisation's people.
- **Changed**: `web/src/App.tsx` loses eight hardcoded slugs; the three route
  files that read `demo/content.ts` read the server instead;
  `internal/tenancy` gains the rename and the people of an organisation.
- **Removed**: nothing yet. `web/src/demo/content.ts` still feeds the project
  shell, which the project settings epic replaces.

## Assumptions

- **One request answers "who am I and where".** The shell needs your name,
  your organisations and which one you are in on every screen. Three requests
  per navigation would be three chances to half-load; one is asked once and
  shared through a context.

- **The address is the truth.** When the address names an organisation you
  are not in, the answer is that there is nothing there, not a redirect to
  one you are in. A redirect would make a wrong link look like a working one
  and hide the mistake.

- **Switching organisations is a navigation, not a mode.** The address
  changes, so the back button works and a link can be sent to somebody.

- **The people of an organisation are its owner's business.** Administering
  the installation puts somebody in or takes them out, which `chatcms-rbdb`
  already allows; the organisation's own owner manages it day to day. Both go
  through the same policy question.

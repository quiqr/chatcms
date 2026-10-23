## Why

Somebody editing needs to see what a reader will see, and see it change as
they type. Epic `chatcms-ihey`, and acceptance test 5.

## What Changes

- **A Hugo server per site**, started when the preview is asked for, using the
  same pinned Hugo that builds use.
- **Reached through the app's own address**, so that a customer's unpublished
  work cannot be read by anybody who may not see the project.
- **Stopped when nobody is looking.** An idle Hugo server is a process holding
  memory for nothing.
- **A cap on how many run at once**, so one busy afternoon cannot take the
  machine down.
- **The preview column in the editor**, with the address, a desktop and phone
  width, and a refresh.

## Capabilities

### New Capabilities

- `preview/serving`: how a preview is reached and who may see one.
- `preview/lifetime`: when a preview starts, when it stops, and how many run.

### Modified Capabilities

None.

## Impact

- **New code**: `internal/preview`, `web/src/routes/Preview.tsx`.
- **Changed**: the editor's preview column holds the real preview; the server
  stops every preview on the way out.
- **Settings**: how long an idle preview is kept, and how many run at once.

## Assumptions

- **Hugo binds only this machine, never the network.** The app is what asks
  who somebody is; a preview listening on its own would be a customer's
  unpublished work with no sign-in in front of it. The app forwards to it.

- **Anybody who may read the project may preview it.** Looking is not
  changing, and a viewer who cannot see the site they are reviewing is a
  viewer who cannot do the one thing the role is for.

- **A preview whose working copy moved is stopped and started again.** The old
  one is serving files that are no longer there, and serving a customer stale
  pages is worse than a moment's wait.

- **Changes show because Hugo watches the files.** Nothing in ChatCMS tells
  the preview to rebuild; it is already watching the working copy that saving
  writes to. That is why a save appears without the preview being restarted,
  and the test waits for it to show rather than sleeping for a guess.

- **The idle check is one loop, not a timer per preview.** One thing to stop
  when the app stops, and no timer can outlive the thing it was counting for.

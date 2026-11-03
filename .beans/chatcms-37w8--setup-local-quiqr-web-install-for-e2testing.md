---
# chatcms-37w8
title: 'setup local quiqr web install for e2testing '
status: completed
type: task
priority: normal
created_at: 2026-11-03T19:52:35Z
updated_at: 2026-11-03T20:17:21Z
openspec-link: openspec/changes/archive/2026-11-03-add-reference-instance
parent: chatcms-0fml
---

make a script to start https://github.com/quiqr/quiqr-desktop as webversion so you can use as reference for testing quiqr sites and advanced models. Use the flake from the project. You might need the fork from mipmip. Check what is the difference. The upstream is prefered.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-03-add-reference-instance`.

### What was built

`bash scripts/reference.sh up` brings the old Quiqr up on a machine that has
none of it: it clones `quiqr/quiqr-desktop` and the kitchen sink template at
the commits pinned in `scripts/reference-pins.env`, builds the standalone
edition inside that checkout's own `nix develop`, and mounts the template as a
site. `stop`, `fresh`, `fields` and `record` do the rest. Everything lives
under `.reference-state/`, which is ignored.

`record` makes the edit itself, through the old application's own form, with
Playwright. The form components are the part whose behaviour is in question,
so an API call would have stepped over them.

### What the bean asked, and what the answer turned out to be

- **Use the flake from the project**: right, and it matters more than it
  looks. That dev shell sets `EMBGIT_PATH` and builds embgit from
  `pkg-embgit.nix`. Embgit is how the old application talks to git, so
  bringing our own node would have meant building embgit ourselves.
- **The fork from mipmip, and what the difference is**: exactly one commit,
  five lines in `flake.nix`, putting libstdc++ on `LD_LIBRARY_PATH` for the
  prebuilt native modules. `scripts/reference.sh` exports that itself, so
  upstream is preferred and nothing is lost.

### What it found, which is what it is for

The first recording said two things nobody knew:

- **The old Quiqr writes a hidden field's default into the file on any save.**
  Named as a divergence in `case.json` with its reason. ChatCMS does not put a
  value in a repository that nobody typed.
- **A document whose front matter ends with a blank line does not survive a
  no-op round trip**: the blank line goes and two newlines arrive at the end
  of the body. The old Quiqr does the same two things, so the recorded case
  passes and a customer's repository gets the same bytes either way, but
  `content/serialisation` claims a no-op round trip changes nothing and that
  is not true for a real file from a real template. Recorded as
  `chatcms-8d0w` rather than fixed here.

### Along the way

- The old application reads `~/Quiqr` unless `storage.dataFolder` says
  otherwise, so the first version of this script listed somebody's real sites.
  It now has a library of its own and its own `HOME`.
- A site is mounted as `<library>/sites/<key>/<workspace>`, not by an absolute
  path in `config.json`.
- Its server answers "Not Found" to a browser asking for a form's address
  directly, so the recording clicks its way in, which is also what a person
  does.

### Proved by

A run from nothing (`fresh`, then the recording again) on a machine with no
checkout: 1.6 seconds for a second `up`, the port free after `stop`, the
repository clean throughout. `TestRecordedEditsMatchTheReference` passes in
the gate, and was made to fail on purpose twice: once by changing a neighbour
in the recorded file, once by removing the named divergence.

One field type is covered: `string`. `docs/TESTING.md` says so plainly, so a
green gate is not read as more than it is.

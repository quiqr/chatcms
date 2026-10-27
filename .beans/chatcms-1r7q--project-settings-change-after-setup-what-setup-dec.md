---
# chatcms-1r7q
title: 'Project settings: change after setup what setup decided once'
status: completed
type: epic
priority: high
created_at: 2026-10-26T16:46:24Z
openspec-link: openspec/changes/archive/2026-10-27-add-project-settings
updated_at: 2026-10-27T00:07:56Z
parent: chatcms-8vem
---

## Purpose

`/p/:project/settings` is a heading with nothing under it. Everything decided
when a project was set up, where the site lives, how it is published, what it
is called, can only be decided once: the setup screen is reachable only while
a project has no site. An implementer who picked the wrong branch, or a
customer who moved host, has no way to change it.

## Scope

- The project settings screen: name, the address shown to customers, and the
  slug, with what happens to existing links when a slug changes said plainly.
- Where the site lives: the address in git, and the key, with "test the
  connection" and a way to make a new key if the old one was removed on the
  host. The same panels the setup screen uses, reused rather than copied.
- How it is published: main line or a line of its own, source files or built
  pages, changed after the fact. Changing it says what the next publish will
  do differently before it is saved.
- Technical details for implementers and owners, behind
  `policy.ActionSeeTechnicalDetail` as everywhere else, and never shown to an
  editor.
- The Publish button's count comes from the project rather than from
  `web/src/demo/content.ts`. It is wrong today whenever the real number
  differs, which `docs/design/COMPARISON.md` records as the first thing still
  to do.
- Removing a project, with a confirmation that names what is removed and what
  is not. The site in git is not touched, and the screen says so.

## Done when

- Every value the setup screen asks for can be changed afterwards and the next
  publish honours it, proven against a real remote on disk.
- Changing the key and testing the connection reports what happened in
  editor-facing words, with the raw output only where technical detail is
  allowed.
- An editor who opens the settings address is refused, by the policy and not
  by a hidden link.
- The count on the Publish button matches the publish screen, and no shell
  reads project data from the demo scaffold.
- Removing a project removes it from ChatCMS and leaves the git remote alone.

## References

- `docs/BRIEFING-M01.md` section 4, projects and publishing.
- `docs/design/reference/screenshots/desktop/onboarding.png`, the panels.
- `docs/design/COMPARISON.md`, "Still to do".
- `internal/publish`, which reads the publishing choice today.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-27-add-project-settings`.

What setting a project up decided once can be changed afterwards: what it is
called, where its site lives, and how it is published.

- **The panels are shared, not copied.** `KeyPanel` and `PublishingPanel`
  live in one place and both the setup screen and the settings screen compose
  them, which is what the bean asked for and what keeps the two from
  drifting.
- **The address a project is reached at does not change.** Renaming is safe;
  moving the address would break every link already sent, and the screen says
  so rather than offering a change that quietly breaks things.
- **Removing a project leaves the site alone**, and the confirmation says so
  in as many words.

### The count was worse than recorded

`docs/design/COMPARISON.md` said the count on the Publish button came from
the scaffold. It was worse: the server never filled the field at all, so it
had always been zero, and the screens hid that by showing a fixed four from
`demo/content.ts`. Reading the real answer made the two disagree out loud.

Both now read one answer, and it counts the same things the publish review
lists rather than versions ahead of the remote: somebody reading "Publish 2"
beside "Ready to publish 2 changes" expects one number.

### The scaffold is gone

`web/src/demo/content.ts` is deleted. Nothing in the interface invents a
project, an organisation or a person any more.

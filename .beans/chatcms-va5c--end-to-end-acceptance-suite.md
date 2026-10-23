---
# chatcms-va5c
title: End-to-end acceptance suite
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T21:25:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-acceptance-suite
parent: chatcms-8vem
blocked_by:
    - chatcms-5uu2
    - chatcms-0aeg
    - chatcms-zytv
    - chatcms-ihey
    - chatcms-ac9e
    - chatcms-tp3b
---

## Purpose

Prove the PoC works. Every acceptance test from the briefing, automated, run
against the built binary, a real database, a real remote and real fixture
sites.

## Scope

- `test/e2e/` with Playwright, one spec per acceptance test from
  `docs/BRIEFING-M01.md` section 6.
- Every spec runs twice, at 1440 px and at 390 px. The phone pass also asserts
  no horizontal page scrolling and no touch target under 44 px.
- A reproducible harness: build the binary, start Postgres, create a bare
  remote, seed fixture sites, run, tear down.
- The `e2e` check in `flake.nix`, as a NixOS VM test if the browser cannot run
  in the sandbox.
- A run against the real customer site, scripted and repeatable, documented in
  the proposal.
- Screenshot comparison against `docs/design/reference/screenshots/` for each
  screen, with differences listed rather than silently accepted.

## Done when

- All nine acceptance tests pass automatically in `nix flake check`.
- The suite is green three times in a row with no flake, and uses no sleeps.
- A deliberately broken behaviour makes the matching spec fail.

## References

- `docs/BRIEFING-M01.md` section 6.
- `docs/TESTING.md`, the end-to-end section.

## Summary of Changes

The nine acceptance tests from `docs/BRIEFING-M01.md` section 6 now run by
themselves, in the gate, against the binary this flake builds.

### What was built

- `test/e2e/`: nine specs, one per acceptance test, plus a harness that gives
  each spec its own PostgreSQL, its own bare remote, its own seeded site and
  its own server, and removes all of it afterwards.
- Every spec runs twice, at 1440 px and at 390 px. The phone pass adds the two
  rules from the design: no sideways scrolling, and nothing to press under
  44 px. Acceptance test 9 is therefore every screen the other eight walk,
  plus a spec that visits the few they do not land on.
- Nothing sleeps. Every wait is for the thing itself, with a deadline: the
  server answering, a save's response, words appearing in the preview frame.
- The `e2e` check in `flake.nix`, with git, Hugo, PostgreSQL and the browser
  all pinned, and no network. It runs in 1.4 minutes.
- `scripts/acceptance-against.sh` runs the same specs against a real
  repository instead of the fixture, which is how the run against the
  customer's own site is done before a release.
- A screen capture at both widths, and `docs/design/COMPARISON.md`, which says
  what differs from the drawings in `docs/design/reference/screenshots/` and
  what was decided about each difference.

### What it proved

- Green three times in a row, and green in the Nix sandbox, with no flake.
- Taking the content update out of going back made spec 08 fail at both
  widths and left the other 25 passing, which is the test the bean asks for.

### What it found, each fixed in the product

- The editor route did not match a page whose key holds a slash, so every
  page below the top level was a not-found.
- The sync screen threw away the answer from bringing changes in, so it still
  said there was work to fetch after fetching it.
- Preview links and live reload pointed at the port Hugo listens on, which
  nothing outside the machine can reach.
- Two controls were under 44 px on a phone, and the phone editor had no way
  to show a preview at all.
- A master key written as raw random bytes was refused whenever its last byte
  happened to be a newline. One start in a hundred, on a machine that had
  worked the day before.
- The integration harness never stopped the PostgreSQL it started, so
  `go test -tags=integration ./...` left one cluster per package running until
  the machine was restarted.
- Chromium in the Nix sandbox has no fonts and no fontconfig, and dies the
  moment it has to pick one. The page still draws, so what it looks like is a
  screen that ignores every keystroke.

### What is left, written down rather than quietly dropped

`docs/design/COMPARISON.md` lists what the screens still do not do that the
drawings show: the count on the Publish button comes from the scaffold rather
than from the project, the projects list has no people or last-edited line,
People has no organisation switch, Media cannot filter by collection. None of
them is an acceptance test; all of them are worth a bean in milestone 02.

---
# chatcms-tp3b
title: Content management
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T16:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-content-management
parent: chatcms-8vem
blocked_by:
    - chatcms-sbe1
    - chatcms-gw8g
---

## Purpose

Browse and manage the pages of a site: the content browser and the create,
rename and delete operations behind it.

## Scope

- `internal/content`: list collections and singles from the model, create,
  rename and delete pages, with the file naming and slug rules the old Quiqr
  uses.
- The content browser from DESIGN.md section 5.5: second-level navigation from
  the model, dots on items with unpublished changes, search, the row menu
  (rename, duplicate, view on site, delete) and "Show N more" pagination.
- The phone content screen with the scrollable collection chip row.
- Every mutation auto-commits through `internal/gitsync` and queues a sync job.
- Deleting asks for confirmation and says what happens.

## Done when

- Create, rename, duplicate and delete all produce the file layout the old
  Quiqr produces, proven against fixtures.
- Unpublished change indicators come from git state, not from a flag the app
  sets by hand.
- Acceptance test 4 from the briefing passes for the page half.

## References

- `docs/BRIEFING-M01.md` section 4.5.
- `docs/design/DESIGN.md` section 5.5.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-content-management`.

A content creator can now see what their site holds and make, rename, copy and
remove pages. The page half of acceptance test 4 passes, proven by an
integration test that walks the whole life of a page: make it, rename it, copy
it, check the browser reflects each step, then remove it.

- A file is named after its title the way the old Quiqr names it, accents
  folded. Renaming moves the file and rewrites the title inside it, because to
  a content creator those are one act.
- A title that would become a path is folded rather than refused: "A/B test"
  becomes `a-b-test`, which is what somebody meant. A title with no letters or
  digits is refused, because nothing would be left to name the file with.
- The mark on an item waiting to be published comes from the same question the
  publish review asks, so the two cannot disagree.
- Removing is a save, not a second kind of deletion, so the confirmation can
  honestly say the page can be brought back.
- A page the model names cannot be renamed or removed here; its name is part
  of how the site is built.

**The New project button works.** It had been drawn since the design
foundation and wired to nothing: the endpoint shipped with tenancy, the setup
screen with onboarding, and nobody's epic owned the step between, so the only
way to get a project was the development seed. The editor's Publish button was
inert in the same way and is now a link.

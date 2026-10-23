---
# chatcms-tp3b
title: Content management
status: in-progress
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T14:53:37Z
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

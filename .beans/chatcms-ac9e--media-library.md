---
# chatcms-ac9e
title: Media library
status: todo
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T15:32:31Z
parent: chatcms-8vem
blocked_by:
    - chatcms-gw8g
    - chatcms-fdo0
---

## Purpose

Upload, browse and pick images, with alt text and usage information, stored in
git alongside the content.

## Scope

- `internal/media`: upload with content type and size validation, thumbnail
  generation, listing, delete, alt text, and a usage lookup that answers "used
  on which pages".
- Media files land in the site's media directory where the old Quiqr expects
  them, and are committed like any other content change.
- The media library from DESIGN.md section 5.7: filter chips including
  "Not used anywhere", search, drop zone, thumbnail grid, uploading tile with
  progress, and the details panel.
- Picker mode, opened from an image field in the form editor, with
  "Use this image".

## Done when

- Upload, pick, replace, alt text and delete all work on desktop and at 390 px.
- A deleted image that is still in use warns before deleting.
- Thumbnails are cached and rebuildable from git alone.
- Acceptance test 4 from the briefing passes for the image half.

## References

- `docs/BRIEFING-M01.md` section 4.6.
- `docs/design/DESIGN.md` section 5.7.

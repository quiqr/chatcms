---
# chatcms-ac9e
title: Media library
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T18:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-media-library
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-media-library`.

The image half of acceptance test 4 passes. A site's pictures can be added,
searched, described, picked from a form field and removed.

- **A file's name is not evidence.** What a picture is comes from reading its
  bytes; the name is then rewritten from what they turned out to be. A file
  called `.png` that is a text document would otherwise be published as a
  broken picture on a customer's site. SVG is recognised by its own opening
  tag, since sniffing reports it as plain text.
- **Nothing is replaced.** A second picture of the same name is put beside the
  first, because replacing would silently change every page using it.
- **Descriptions live in `quiqr/media.yaml`**, written with names in order so
  the same descriptions always make the same file and a publish shows no
  spurious change.
- **What uses a picture is a search, not an index.** An index would be wrong
  the first time anybody edited a page outside ChatCMS, which this product
  supports on purpose. Every shape a Hugo site writes an address in is looked
  for.
- **A small copy is a cache in the strict sense**, named after a hash of its
  source bytes: a changed picture gets a new name by itself, and the whole
  directory can be deleted at any moment and rebuilt from git alone.
- Removing a picture something uses says how many pages would have a gap, and
  needs asking twice.

One dependency added: `golang.org/x/image`. The standard library decodes
pictures but cannot resize one.

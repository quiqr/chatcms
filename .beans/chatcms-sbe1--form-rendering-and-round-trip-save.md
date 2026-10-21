---
# chatcms-sbe1
title: Form rendering and round-trip save
status: in-progress
type: epic
priority: critical
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-21T20:37:36Z
parent: chatcms-8vem
blocked_by:
    - chatcms-vho6
---

## Purpose

Render forms from the model and write the result back identically to the old
Quiqr. This epic carries the backwards-compatibility contract, so it is built
golden tests first.

## Scope

- `internal/forms`: load a content file into form values, and write form values
  back to front matter plus Markdown body, preserving key order, quoting style,
  date formats, list indentation and the body separator.
- Field renderers in `web/src/fields/` for text, date, select, image, textarea,
  Markdown body, list of items and collapsed groups, matching DESIGN.md
  section 5.6.
- A Markdown body editor with the toolbar from the design and a raw source
  toggle, defaulting to formatted editing.
- Field to preview linking: the focused field is outlined in the preview.
- The golden test harness in `test/fixtures/`, with goldens generated from the
  reference app, plus the `golden` check in `flake.nix`.
- Autosave with the sync indicator states, no Save button anywhere.

## Done when

- A load-and-save round trip with no edits is byte identical to the old Quiqr
  output for every fixture.
- A known edit per field type matches its golden.
- `internal/forms` is at or above 80% statement coverage.
- Acceptance test 3 from the briefing passes.

## References

- `docs/TESTING.md`, the golden tests section.
- `docs/design/DESIGN.md` section 5.6.

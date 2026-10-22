---
# chatcms-fdo0
title: 'Design foundation: tokens, components, shells'
status: in-progress
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-22T12:47:10Z
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

Build the visual foundation so every later screen is assembled from existing
parts rather than invented. The design is the source of truth for look; the
spec is the source of truth for behaviour.

## Scope

- Merge `docs/design/tokens/tokens.css` into the global stylesheet, keeping
  shadcn variable names. Load Geist and Geist Mono. Copy
  `docs/design/assets/quiqr-logo.svg` into `web/public/`.
- Add the shadcn components listed in DESIGN.md section 4 and restyle the
  variants: primary, ink/secondary, outline, ghost, status badges, segmented
  control, radio cards, avatars, project tiles.
- Build the three shells from DESIGN.md section 3: the project shell with
  sidebar and top bar, the phone shell with bottom tab bar, and the
  organisation header.
- Reserve the editor's three-slot flex row and the 64 px icon rail so the
  assistant panel can land later without a rewrite.
- Set up routing and dummy data for every milestone 01 screen.
- Add the UI string lint for the banned git words from DESIGN.md section 7, and
  wire it into the frontend check.

## Done when

- The shell renders at 1440 px and at 390 px matching
  `docs/design/reference/screenshots/desktop/content.png` and
  `phone/phone-content.png`, with any remaining difference listed in the
  proposal.
- No horizontal page scrolling at 390 px and no touch target under 44 px.
- The banned-word lint fails on a deliberately introduced "commit" string.
- No hard-coded hex values in components.

## References

- `docs/design/DESIGN.md` sections 1 to 4 and 7.
- `docs/design/PROMPT.md` for the intended build order.

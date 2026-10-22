---
# chatcms-fdo0
title: 'Design foundation: tokens, components, shells'
status: completed
type: epic
priority: high
created_at: 2026-10-21T15:32:21Z
openspec-link: openspec/changes/archive/2026-10-22-add-design-foundation
updated_at: 2026-10-22T13:04:44Z
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

## Summary of Changes

Shipped as OpenSpec change `add-design-foundation`, archived at `openspec/changes/archive/2026-10-22-add-design-foundation`. Three capabilities specified and built: `interface/design-system`, `interface/app-shell` and `interface/editor-language`, plus a requirement added to `platform/frontend-delivery`.

The tokens from the hand-off are the only source of colour, radius and type. The two typefaces are installed as packages and served by the application rather than fetched from Google, so no reader contacts a third party and the build needs no network.

The components the designs call for exist with the variants they use: buttons in four styles with a count chip, status badges with dots where the design draws them, the segmented control, radio cards, menus, dialogs, disclosures, toasts, tables, tooltips, avatars and project tiles. 37 tests cover them and the shells.

Three frames: the project shell, the phone shell with its bottom tab bar padded clear of the device, and the organisation header. Both frames are rendered and the stylesheet picks one, so there is nothing to measure and nothing to flash. The editor frame is a row of slots with a narrow rail and an empty third slot, so the assistant panel lands later without a rebuild. All ten milestone screens have an address.

Three rules are enforced by the build rather than by review: no colour written by hand, no control below its minimum size, and no version control word in a string a content creator can read. Each was proven to fail on a deliberate breach, and the audience marker that allows technical words in implementer-only files was proven to work. The default is the strict one, so forgetting the marker makes the check stricter rather than looser.

Two things worth recording. The hand-off contradicts itself on the segmented control: section 2 sets a thirty-six pixel floor and the drawn control is thirty-two. The hand-off says to follow the screen for visuals and the spec for behaviour and to mention the conflict, so the control is thirty-six here. And comparing the render against the reference screenshots by hand found six differences; five were fixed and the sixth is recorded in the proposal.

A real browser confirmed at 390 pixels that no page scrolls sideways and no control is under forty-four pixels.

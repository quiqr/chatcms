## Why

Every screen in milestone 1 is assembled from the same parts: the same
colours, the same buttons, the same shell around them. Building those parts
once, from the design hand-off, is what stops ten screens becoming ten
dialects. Epic `chatcms-fdo0`.

It serves acceptance test 9, "everything works on a phone", and it unblocks
the second half of the form epic, which is waiting for components to render
fields with.

## What Changes

- **The design tokens become the only source of colour and type.** The tokens
  from `docs/design/tokens/tokens.css` are merged into the application's
  stylesheet under the names shadcn uses, and a check fails the build when a
  component writes a colour of its own.
- **The typefaces ship with the application.** Geist and Geist Mono are served
  from the application rather than fetched from a third party, so no reader's
  browser contacts anyone else and the build needs no network.
- **The components the designs use exist.** The shadcn components listed in
  `docs/design/DESIGN.md` section 4, with their variants restyled to the
  designs: primary, ink, outline and ghost buttons, status badges, the
  segmented control, radio cards, menus, dialogs, disclosures, toasts, tables,
  tooltips, avatars and project tiles.
- **The three shells exist.** The project shell with its sidebar and top bar,
  the phone shell with its bottom tab bar, and the organisation header, each
  reflowing the way section 3 describes.
- **The editor's layout leaves room for what comes later.** A three-slot row
  and a narrow icon rail, so the assistant panel lands in milestone 2 without
  the editor being rebuilt.
- **Every screen has a route and something to show.** Routes for all ten
  milestone 1 screens, with clearly marked placeholder data, so each screen
  epic fills in a page rather than inventing navigation.
- **Editor-facing words are checked, not trusted.** A check fails the build
  when a word from the banned list in section 7 reaches a string a content
  creator can read.

## Capabilities

### New Capabilities

- `interface/design-system`: what the interface is made of, and the rules that
  keep it one system rather than ten.
- `interface/app-shell`: the frames every screen sits in, and how they behave
  as the window narrows.
- `interface/editor-language`: which words may reach a content creator, and
  where technical detail is allowed instead.

### Modified Capabilities

- `platform/frontend-delivery`: the frontend check grows the two new rules, so
  a colour written by hand or a git word in an editor's view fails the gate
  rather than review.

## Impact

- **New code**: `web/src/components/`, `web/src/shells/`, `web/src/routes/`,
  `web/src/demo/`.
- **Changed code**: the global stylesheet, the application entry point, and
  `web/package.json`.
- **Dependencies added**: the Radix primitives the listed components need,
  `class-variance-authority`, `clsx`, `tailwind-merge`, `lucide-react`,
  `react-router`, and the two Geist typefaces as packages. All pinned, and the
  pinned `npmDepsHash` in `flake.nix` moves with them.
- **Gate**: two checks are added to the frontend check and both are proven to
  fail.
- **Seams**: none. This change has no backend.

## Assumptions

- **The typefaces are vendored rather than linked.** The tokens file links
  Google Fonts. A hosted CMS whose customers are Dutch should not make every
  reader's browser contact a third party, and a build that fetches at runtime
  cannot be checked offline. The same two typefaces are installed as packages
  and served by the application. Nothing about the design changes.
- **Components are copied in, not depended on.** That is how shadcn works:
  its components are source files in the repository, which is also what makes
  the Nix build need no network.
- **Pixel comparison against the reference screenshots waits for the
  end-to-end epic.** This change checks what can be checked without a browser:
  that tokens are used, that no colour is written by hand, that the shells
  reflow by the rules the design states, and that no banned word is reachable.
  Comparing rendered pixels needs a browser in the gate, which the end-to-end
  epic sets up.

  The result was nonetheless rendered in a real browser at 1440 and 390 and
  compared with `reference/screenshots/desktop/content.png` and
  `phone/phone-content.png` by hand. Six differences were found and five fixed:
  the word "ChatCMS" was missing beside the logo; the project switcher was a
  plain row rather than a bordered card with a switch affordance; the
  navigation carried uppercase group headings the designs do not have; the
  unpublished count was missing from the sidebar and from the phone tab bar;
  and the sync indicator coloured its text rather than only its icon. The one
  left is the active navigation item, which the designs give a faint border
  and this build gives only a tinted background; it is listed here rather than
  fixed because the content screen's own epic will restyle that row anyway.

  The browser run also confirmed, at 390 pixels across the three frames, that
  no page scrolls sideways and that no control is below forty-four pixels.
- **Screens are not built here.** Each screen belongs to the epic that owns
  its behaviour. This change gives each one a route and a frame.
- **One conflict in the hand-off is resolved in favour of the rule.** Section
  2 of `docs/design/DESIGN.md` says a secondary control on a pointer device
  may be as small as thirty-six pixels, and the drawn segmented control in
  `reference/html/editor.html` is thirty-two. The hand-off itself says to
  follow the screen for visuals and the spec for behaviour, and to mention the
  conflict: a hit target is behaviour, so the control is thirty-six pixels
  here and four pixels taller than drawn.

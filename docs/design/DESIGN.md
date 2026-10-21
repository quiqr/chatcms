# Quiqr ChatCMS — UI spec (milestone 1)

This is the design spec for the hosted Quiqr ChatCMS web app. Build it in
React + Tailwind (v4) + shadcn/ui. The reference screens in
`reference/html/` and `reference/screenshots/` are the source of truth for
look and layout; this file explains the system, the behaviour and the
reasoning so you can generalise beyond the drawn screens.

When this spec and a reference screen disagree, follow the reference screen
for visuals and this spec for behaviour, and mention the conflict.

---

## 1. Principles

1. **Calm and functional.** Neutral surfaces, one blue accent, no gradients,
   no emoji, no decorative illustrations. Cards only where something is a
   separate object.
2. **Plain language for content creators.** Editors never see git terms:
   no commit, branch, push, pull, merge, repo, SHA, build log. Use
   *saved*, *changes*, *publish*, *version*, *go back to*.
3. **Technical details only for implementers.** Owners and implementers can
   reveal technical information (deploy key, branch, build log, sync
   internals) in clearly secondary places: a "Technical details" disclosure,
   a hover/tap tooltip, the Setup section. Editors and viewers never see it.
4. **Saving is automatic; publishing is deliberate.** No Save buttons.
   Publish is the one action that gets a review step.
5. **Mobile is a first-class target.** Every screen works at 390 px. Desktop
   layouts reflow to one column; the core editor flow has a dedicated phone
   layout (bottom tab bar, sticky Publish).
6. **Errors say what happened, that nothing is lost, and what to do next.**

## 2. Foundations

Tokens live in `tokens/tokens.css` (shadcn variable names + extras). Use the
Tailwind utilities they generate (`bg-card`, `text-muted-foreground`,
`bg-warning-soft text-warning-foreground`, …). Never hard-code hex in
components.

**Type** — Geist (UI), Geist Mono (keys, branch names, logs, Markdown toggle).

| Use | Size / weight |
| --- | --- |
| Page title (Publish, Projects) | 28–30 px / 650, tracking −0.025em |
| Section title (inside screens) | 22 px / 650 |
| Top-bar title | 18 px / 600 |
| Card title, field group title | 15–16 px / 600 |
| Body, inputs | 15 px (16 px on phone inputs to avoid iOS zoom) |
| Labels | 14 px / 500 |
| Meta, captions, badges | 12–13 px |

**Spacing** — 4 px base. Typical: field gap 22 px, card padding 16–24 px,
page padding 24–32 px desktop / 16 px phone. Lay out siblings with flex/grid
`gap`, not margins.

**Radius** — inputs & buttons 8 px, small chips 6 px, list cards 12–14 px,
dialogs 16 px, pills/badges 999 px.

**Elevation** — borders, not shadows. Shadows only on overlays: menus and
popovers `0 12px 32px rgb(16 24 40 / .12)`, dialogs `0 24px 64px rgb(16 24 40 / .24)`,
the preview frame a soft `0 12px 32px rgb(16 24 40 / .08)`.

**Hit targets** — at least 44 px on touch; 36–40 px is acceptable for
secondary desktop controls.

**Accessibility** — real `<button>`, `<a>`, `<label>`; `aria-label` on
icon-only buttons; `aria-current="page"` on active nav; visible focus ring
(`ring` colour, 2 px + 3 px `primary-soft` halo on inputs). Text ≥ 4.5:1.
The logo colours (`quiqr-*`) are never used for text on white.

**Icons** — lucide-react, 1.8 px stroke, 16–20 px. Mapping used in the
designs: Content `FileText`/`File`, Media `Image`, Publish `Upload`,
History `History`, People `Users`, Settings `SlidersHorizontal`,
Workflows `Workflow`, Assistant `MessageSquare`, Saved `CloudCheck`,
Saving/sync `RefreshCw`, Offline `CloudOff`, Preview `Eye`.

**Logo** — `assets/quiqr-logo.svg` (official multicolour Quiqr wordmark)
followed by "ChatCMS" in `muted-foreground`. Height 24–28 px in headers.

## 3. App shell

### Project shell (all screens inside a project)
- **Sidebar** (240 px, `bg-card`, right border): logo · project switcher
  (project tile + name + domain, opens the project list) · nav group
  *Site*: Content, Media, Publish (with count badge of unpublished
  changes), History · nav group *Setup* (only owner/implementer): People &
  roles, Project settings, Workflows ("Soon" badge) · spacer · current user
  (avatar, name, role).
- **Top bar** (`bg-card`, bottom border, 12 px × 32 px padding): title or
  breadcrumb · sync indicator · page actions · primary **Publish** button
  with count.
- **Responsive**: shell is `flex flex-wrap`; content is
  `flex: 999 1 560px; min-width: 0`, so under ~800 px the sidebar stacks
  above content. On phones replace the sidebar with the phone shell below.
- **Editor exception**: the page editor uses a 64 px icon rail instead of
  the full sidebar to give room to form + preview.

### Phone shell (< 640 px)
- Header: project tile, project name, sync dot + text, avatar. Under it a
  horizontally scrollable chip row for collections where relevant.
- Bottom tab bar: Content, Media, Publish (badge), History. 52 px targets,
  padded by `env(safe-area-inset-bottom)`.
- Editor and Publish hide the tab bar and show a sticky bottom action bar
  (Assistant icon button + full-width Publish).
- Don't draw a fake status bar; use safe-area insets.

### Organisation-level pages (project overview)
Top header instead of sidebar: logo / organisation switcher (dropdown with
organisations + "New organisation") / Help / avatar. Content max-width
1200 px.

## 4. Components (map to shadcn/ui)

| Pattern | shadcn base | Notes |
| --- | --- | --- |
| Primary button | `Button` | `bg-primary`, 38–44 px high, 600 weight. Publish carries a count chip (`bg-white/20`). |
| Ink button | `Button variant="secondary"` restyled | `bg-secondary text-secondary-foreground`. For create/upload actions so they don't compete with Publish. |
| Outline / ghost | `Button variant="outline" / "ghost"` | Preview site, Back, Replace, row actions. |
| Status badge | `Badge` | Soft bg + dark text: Changed = warning, New = info, Published = plain muted text, Live now/Connected = success with dot, Needs attention = destructive with dot. |
| Segmented control | `ToggleGroup` / `Tabs` | `bg-muted` track, white selected pill with tiny shadow. Form / Side by side / Preview; This project / Organisation; source tabs in onboarding. |
| Radio cards | `RadioGroup` | Bordered tiles; selected = 1.5 px `primary` border + `primary-tint` bg. |
| Menus | `DropdownMenu` | Row actions (Rename, Duplicate, View on site, Delete… in destructive colour, separated). Org switcher. |
| Dialog | `AlertDialog` | Rollback confirmation, delete confirmation. |
| Disclosure | `Collapsible` | "Technical details", optional field groups ("Search engines & sharing"). |
| Toast | `Sonner` | Dark toast for "Joris updated 'About us' · Show". |
| Table | `Table` in an `overflow-x-auto` wrapper | People list. |
| Tooltip/Popover | `Tooltip` / `Popover` | Implementer sync details. |

Avatars: initials on a soft tint. Project tiles: first letter on a soft tint.

## 5. Screens

Each section names its reference file (`reference/html/<file>` and
`reference/screenshots/{desktop,phone}/<file>.png`).

### 5.1 Sign in — `login`
Split screen: form left (max 380 px), soft `primary-soft` panel right with a
mini form + preview illustration and the line "Fill in a form, or just ask.
Your site updates as you go." Panel stacks under the form on phones.
Fields: Email, Password (+ "Forgot password?"), "Keep me signed in on this
device", Sign in. Hint: "Got an invitation email? Use the link in it to set
your password first."

### 5.2 Organisation & projects — `projects`
Org switcher dropdown in header. Title "Projects", search, **New project**.
Card grid (`auto-fit, minmax(340px, 1fr)`): tile, name, `domain · Hugo`,
status badge (N unpublished changes / Published / Publishing… / Needs
attention), footer with last activity and people count. Last tile is a
dashed "Import a site or start from a template".

### 5.3 Project onboarding (implementer) — `onboarding`
Left step list (Repository ✓, Deploy key ✓, Publishing — current, Invite
people). Right, three section cards:
1. **Repository** — source tabs (Import Git repository / From Quiqr Desktop /
   Start from a template), repo URL (mono), detection result ("Hugo site ·
   Quiqr model found ✓").
2. **Deploy key** — explanation, read-only key box with Copy, status badge
   "Connected · read & write", links: how-to, Test connection again.
3. **Publishing** (active, primary border + halo) — "Push to": Main branch /
   A separate branch (+ branch name input, default `quiqr/publish`).
   "What to push": Source files (your pipeline builds) / Built HTML (Quiqr
   builds with Hugo). Note that custom build steps will live in
   `quiqr/workflows/` (editor comes later). Back / Continue to invite people.

### 5.4 People & roles — `people`
Top bar segmented control: This project / Organisation. Invite row (email,
role select, Send invite). Table: person (avatar, name, email), role select
(own row shows fixed Owner badge), last active, actions (… menu, or
Resend for pending invites shown with dashed avatar and warning text).
Side card "What each role can do":
- Owner — everything, including billing and deleting the project.
- Implementer — sets up the repository, publishing and people; sees technical details.
- Editor — edits pages and media, and publishes the site.
- Viewer — reads content and previews; can't change anything.

### 5.5 Content browser — `content`
Second-level nav from the Quiqr model: *Pages* (singles) and *Collections*
(with counts); dot on items with unpublished changes. Main: collection
title, search, **New article** (label uses the collection's item name).
List rows: thumbnail, title, "Edited by X · time", status badge, … menu
(Rename, Duplicate, View on site, Delete…). "Show N more" pagination.

### 5.6 Form editor + live preview — `editor`, `editor-assistant`
- Top bar: back, collection + page title, sync status, layout toggle
  (Form / Side by side / Preview), **Assistant**, **Publish**.
- Form column (max 760 px): fields rendered from the model. Shown field
  types: text, date, select, image (thumbnail, file name + size, Replace /
  Remove), textarea with helper + character count, Markdown body (toolbar:
  heading, bold, italic, link, list, image; "Markdown" toggle to raw
  source; default is formatted editing), list of items (drag handle,
  input, remove, "Add <item>"), collapsed optional group.
- The focused field is highlighted in the preview with a dashed primary
  outline (field ↔ preview linking).
- Preview column: address pill "Preview · <url>", device toggle
  (desktop/phone), the real site rendered in a framed panel.
- Assistant (future): third column (~340 px) with chat messages; changes
  the assistant makes are listed as field chips with Show / Undo, and the
  form updates live. In M1 show the button and keep the slot; panel marked
  "Coming later".
- Phone (`phone-editor`): Edit / Preview segmented toggle under header,
  stacked fields, sticky bottom bar with Assistant + Publish.

### 5.7 Media library — `media`
Filter chips (All · count, by section, "Not used anywhere"), search, drop
zone, thumbnail grid (`auto-fill, minmax(150px, 1fr)`), uploading tile with
progress bar. Selecting an image opens a details panel: preview, name,
size, who/when, "Description for screen readers" (alt text), "Used on"
links, **Use this image** (when opened as a picker) and delete icon.

### 5.8 Publish — `publish`, `publish-building`, `publish-done`, `publish-error`
- **Review**: "Ready to publish N changes", last-published line. List of
  changes with badge (Changed / New page / New image / Removed), page name,
  which fields, who. "Show changes" expands an inline word diff (removed:
  red soft strike-through; added: green soft). Optional note "What's in this
  update?" (shown in History). Keep editing / **Publish N changes**.
- **In progress**: four plain steps — Saving your changes, Checking pages and
  links, Building the site (n of m pages + progress bar), Putting it live.
  "You can keep editing."
- **Done**: success icon, "Your site is up to date", time, View site / Back
  to content.
- **Error**: "We couldn't publish this time." + "Nothing changed on your live
  site, and your edits are safe." Specific human cause with a direct action
  ("Open the page"), Try again, "<implementer> has been told about this".
  Collapsible "Technical details · for implementers" with the raw log
  (only rendered for owner/implementer).

### 5.9 History & rollback — `history`, `history-rollback`
Timeline grouped by period. Top item: dashed "Not published yet · N
changes" with Review link. Each version: title (publish note or generated
summary), date · who · N changes, "Live now" badge on the current one.
Selected version → side panel with its changes, "Preview the site as it
was", **Go back to this version…**. Confirmation dialog: what will happen,
"Nothing is lost … you can undo this", warning box if unpublished changes
will be set aside, checkbox "Publish right away" (default on).

### 5.10 Sync status — `sync`
One quiet indicator in the top bar:
- Saved — cloud-check icon (success) + "All changes saved".
- Saving — spinning refresh (info) + "Saving…". Never blocks typing.
- Offline — warning pill "Offline · saved on this device"; syncs on reconnect.
- Updates from others — pulled in automatically; dark toast "Joris updated
  'About us' · Show".
- Same field edited by two people — inline on the field: "Joris also changed
  this summary. Which one should we keep?" Keep mine / Use Joris's / Compare.
- Implementer details on hover/tap: branch, head, remote state (mono).

## 6. Roles and visibility

| | Owner | Implementer | Editor | Viewer |
| --- | --- | --- | --- | --- |
| Content, Media (edit) | ✓ | ✓ | ✓ | read-only |
| Publish, Go back to version | ✓ | ✓ | ✓ | – |
| Setup nav (People, Project settings, Workflows) | ✓ | ✓ | – | – |
| Technical details (logs, branch, key) | ✓ | ✓ | – | – |
| Billing, delete project | ✓ | – | – | – |

Hide what a role can't use; don't show disabled controls for it.

## 7. Copy rules

- Sentence case everywhere. Buttons say exactly what happens ("Publish 4
  changes", "Go back and publish", "Send invite").
- Banned words in editor-facing UI: commit, branch, push, pull, merge,
  repository, deploy, build log, HEAD, SHA, conflict. Use: saved, changes,
  publish, version, go back to, "edited by two people".
- Times are relative and human ("12 min ago", "Tuesday at 9:12").
- Errors: what happened → reassurance → one concrete fix.
- UI language: English in the designs; plan for i18n (Dutch next).

## 8. Room for later milestones

- **Assistant panel**: right-hand column in the editor (desktop) / full-height
  sheet from the Assistant button (phone). Keep the editor layout as a
  3-slot flex row now.
- **Workflows editor**: Setup › Workflows (currently "Soon").
- **Artifacts** (e.g. PDFs from Quarto builds): expected as an extra item in
  the Site nav after History, and as downloadable items on the Publish
  "Done" state.
- **Dark mode**: tokens are named so a dark block can be added later.

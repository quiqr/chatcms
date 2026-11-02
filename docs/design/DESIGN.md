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
3. **Technical details only for implementers.** Implementers and admins can
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
the user's organisations + "Manage <organisation>" for organisation admins) /
"Platform admin" button (platform admins only, ink pill with shield icon) /
Help / avatar. Content max-width 1200 px. Only platform admins create
organisations, so there is no "New organisation" here.

### Organisation admin shell
Light sidebar like the project shell: logo · organisation tile ("Studio Noord ·
Organisation") · nav group *Manage organisation*: Projects, Members (with
counts) · **usage box** (`bg-background` panel: Members n of max, Projects n of
max, each with a 6 px bar; "Limits are set by the platform admin"; when a limit
is off show "n · no limit" without a bar) · spacer · "Back to my projects" ·
user with role "Organisation admin".

### Platform admin shell
Dark sidebar (`bg-secondary`, text `#FAFAFA`, secondary text `#A1A1AA`) so it is
always obvious you are in the admin area: logo · "You're in Platform admin"
panel with shield icon (`#27272A` bg) · nav: Organisations, Users (with counts;
active item `#3F3F46`) · spacer · "Back to my projects" · user with role
"Platform admin". Content area is the normal light page.

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
role select: Editor / Viewer / Implementer, Send invite). Table: person
(avatar, name, email), role select (organisation admins show a fixed
"Organisation admin" badge instead), last active, actions (… menu, or Resend
for pending invites shown with dashed avatar and warning text).
Side card "What each role can do":
- Organisation admin — has access to every project in the organisation; managed under organisation members (§5.12).
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

### 5.11 Roles & hierarchy — `hierarchy`
Reference board (not an app screen): three level cards (platform admin dark,
organisation admin with primary border, project roles), a sentence on roles
per place, and the "limits are hard" rule.

### 5.12 Organisation admin · Members — `org-members`, `org-members-full`
Top bar: "Members" + seat pill ("7 of 10 seats used"; red "All 10 seats in
use" when full). Invite card: Email, Role in organisation (Member /
Organisation admin), Project access (project + role), Send invite; helper
"3 seats left. Invitations that haven't been accepted count as a seat."
Table: person, role in organisation (select; own row fixed badge), projects as
chips ("Korenaar Bakery · Editor", "All projects" for admins), last active,
Remove / Resend + Cancel invite. **Full state**: warning banner "All 10 seats
are in use. Remove someone, or ask the platform admin to raise the limit." +
"Request more seats"; invite fields muted and Send invite disabled.

### 5.13 Organisation admin · Projects — `org-projects` (+ `-duplicate`, `-duplicate-at-limit`, `-move`, `-move-at-limit`)
Top bar: "Projects", limit pill ("4 of 5 · 1 left"), New project. List (not a
table, so it reflows on phones): tile, name, "domain · n people · published
<when>", status badge, Open, … menu: People & roles, Project settings,
Rename | Duplicate…, Move to another organisation…, Download as zip | Archive,
Delete project…. Collapsed "Archived · n · doesn't count toward the limit".
- **Duplicate dialog**: name field prefilled with the original name and
  flagged invalid — the copy **must** get a different name ("This name is
  already taken. Give the copy its own name."). What to copy: pages, media and
  content model (always), people and roles (off), publishing settings (off,
  "so the copy can't overwrite the live website"). Shows the limit impact; at
  the limit it's blocked (§Limits).
- **Move dialog**: destination radio cards. Only organisations where the user
  is organisation admin are selectable; others show greyed with the reason
  ("You're a member here, not an admin"). "What happens": website, publishing
  and history move along; people keep their project roles and become members
  of the destination; the old organisation's admins lose access. Blocked state
  when the destination's member limit would be exceeded.
- **Download as zip**: starts a download of pages, media and the content
  model; no dialog.

### 5.14 Platform admin · Organisations — `admin-organisations`
Top bar: title, search, New organisation. Filter chips: All · n, At a limit,
No admin yet. Table: organisation (tile, name, slug → detail), organisation
admins (or "No admin yet · invite pending" in warning colour), Members and
Projects usage cells (n of max + bar + Almost full/Full, or "n · no limit"),
created, Manage. Footnote explaining optional, hard limits.

### 5.15 Platform admin · Organisation — `admin-organisation`, `admin-organisation-delete`
Header: back, tile, name, Save changes. Cards: **Details** (name, address
`chatcms.eu/<slug>`, created by/when); **Limits** (per limit a checkbox
"Limit members"/"Limit projects" + −/value/+ stepper, usage line and bar;
note "Limits are optional…"); **Projects** (rows with Duplicate, Move, Zip —
platform admins can move to any organisation); **Members** (add by email +
role, table with role select, projects, Remove); **Delete organisation**
(danger card). Delete dialog lists consequences and requires typing the
organisation name.

### 5.16 Platform admin · Users — `admin-users`
Top bar: title, search by name or email, Invite user. Filter chips: All,
Platform admins, Invited, Blocked, Not in any organisation. Table: user (avatar,
name, "Platform admin" ink badge, email), organisations ("Studio Noord
(admin), …"), status badge (Active / Invited / Blocked), last sign-in, … menu:
Edit name and email, Send password reset, Add to organisation…, Make platform
admin | Block sign-in, Delete user…. Pagination "Showing 1–8 of 46".

## 6. Roles and visibility

Three levels; each manages the one below, within limits set from above
(reference: `hierarchy`). Roles belong to a place, not a person: someone can be
organisation admin in one organisation and editor on a project in another.

| | Platform admin | Organisation admin | Implementer | Editor | Viewer |
| --- | --- | --- | --- | --- | --- |
| Scope | whole platform | one organisation | one project | one project | one project |
| Create / edit / delete organisations | ✓ | – | – | – | – |
| Manage all user accounts | ✓ | – | – | – | – |
| Set (optional) member & project limits | ✓ | – | – | – | – |
| Invite / remove organisation members | ✓ (any org) | ✓ | – | – | – |
| Create, import, archive, delete projects | ✓ (any org) | ✓ | – | – | – |
| Duplicate project, download as zip | ✓ | ✓ | – | – | – |
| Move project to another organisation | any org | only orgs they also admin | – | – | – |
| Assign project roles | ✓ | ✓ | ✓ | – | – |
| Project setup (repo, publishing, workflows) | ✓ | ✓ | ✓ | – | – |
| Technical details (logs, branch, key) | ✓ | ✓ | ✓ | – | – |
| Edit content & media, publish, roll back | ✓ | ✓ | ✓ | ✓ | – |
| Read & preview | ✓ | ✓ | ✓ | ✓ | ✓ |

Organisation admins can open every project in their organisation. Members
only see projects they've been given, with that project's role.

Hide what a role can't use; don't show disabled controls for it.

### Limits (rules)
- Both limits (max members, max projects) are **optional** per organisation;
  default for a new organisation is no limit.
- Limits are **hard**: any action that would take an organisation over a limit
  is refused before it runs — inviting, adding a member, creating/importing a
  project, duplicating, moving a project in (counting the people who'd become
  new members). The primary button is disabled and an inline error
  (`destructive-soft` box, ban icon) names the limit, the numbers ("That makes
  6 members, and the limit is 5") and the ways out (remove something, ask the
  platform admin to raise or remove the limit, pick another organisation).
- A pending invitation counts as a seat. Archived projects don't count.
- Lowering a limit below current usage never removes anything; it only blocks
  new additions until usage is under the limit.
- Usage display: bar in `primary` below 80 %, `warning` from 80 %, `destructive`
  at 100 % with "Almost full" / "Full" labels.

## 7. Copy rules

- Sentence case everywhere. Buttons say exactly what happens ("Publish 4
  changes", "Go back and publish", "Send invite").
- Banned words in editor-facing UI: commit, branch, push, pull, merge,
  repository, deploy, build log, HEAD, SHA, conflict. Use: saved, changes,
  publish, version, go back to, "edited by two people".
- Times are relative and human ("12 min ago", "Tuesday at 9:12").
- Errors: what happened → reassurance → one concrete fix.
- UI language: English in the designs. Every string lives in
  `web/src/strings/`, and the interface is readable in English and in Dutch;
  a language is chosen on the preferences screen and kept with the account.
  The copy rules above hold in every language, and the check reads the
  catalogue in each of them.

## 8. Room for later milestones

- **Assistant panel**: right-hand column in the editor (desktop) / full-height
  sheet from the Assistant button (phone). Keep the editor layout as a
  3-slot flex row now.
- **Workflows editor**: Setup › Workflows (currently "Soon").
- **Artifacts** (e.g. PDFs from Quarto builds): expected as an extra item in
  the Site nav after History, and as downloadable items on the Publish
  "Done" state.
- **Dark mode**: tokens are named so a dark block can be added later.

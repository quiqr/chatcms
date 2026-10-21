# Prompts for Claude Code

Adjust `docs/design/` if you put the folder somewhere else.

## Kick-off

```
We're building the frontend of Quiqr ChatCMS (React + Tailwind v4 + shadcn/ui).
The UI design is in docs/design/:
- docs/design/DESIGN.md is the spec. Read it fully first.
- docs/design/reference/screenshots/ has desktop (1440 px) and phone (390 px)
  renders of every screen; docs/design/reference/html/ has the same screens as
  static HTML with exact inline styles.
- docs/design/tokens/tokens.css has the design tokens.

Step 1 only: set up the design foundation.
1. Merge tokens.css into our global CSS (keep shadcn's variable names), load
   Geist and Geist Mono, copy assets/quiqr-logo.svg to public/.
2. Add the shadcn components listed in DESIGN.md §4 and restyle variants to
   match (primary, ink/secondary, outline, ghost; status badges; segmented
   control; radio cards).
3. Build the app shell from §3: project sidebar + top bar with sync indicator
   and Publish button, the phone shell with bottom tab bar, and the
   organisation header. Use dummy data and routes for now.
4. Show me the result at 1440 px and 390 px next to
   reference/screenshots/desktop/content.png and phone/phone-content.png and
   list any differences.

Don't build the screens yet. Follow DESIGN.md's copy rules: no git terms in
editor-facing UI.
```

## Then, one screen at a time

```
Build the <SCREEN> screen following DESIGN.md §5.<N>.
Reference: docs/design/reference/screenshots/desktop/<file>.png,
docs/design/reference/screenshots/phone/<file>.png and
docs/design/reference/html/<file>.html (exact spacing/colours in inline styles).
Use our tokens and shadcn components, not the inline hex values.
Wire it to <API / dummy data>. Respect the role rules in §6.
When done, screenshot it at 1440 and 390 px and compare with the references.
```

Suggested order: Sign in → Projects → Content browser → Form editor + preview
→ Media → Publish (all four states) → History & rollback → Sync states →
People & roles → Onboarding.

| Screen | §  | Reference file(s) |
| --- | --- | --- |
| Sign in | 5.1 | `login` |
| Organisation & projects | 5.2 | `projects` |
| Project onboarding | 5.3 | `onboarding` |
| People & roles | 5.4 | `people` |
| Content browser | 5.5 | `content`, `phone-content` |
| Form editor + preview | 5.6 | `editor`, `editor-assistant`, `phone-editor` |
| Media library | 5.7 | `media` |
| Publish | 5.8 | `publish`, `publish-building`, `publish-done`, `publish-error`, `phone-publish` |
| History & rollback | 5.9 | `history`, `history-rollback` |
| Sync status | 5.10 | `sync` |

## Useful follow-ups

- "Check every screen at 390 px for horizontal scrolling and touch targets under 44 px."
- "Grep the UI strings for banned git words from DESIGN.md §7 and fix them."
- "Make the form editor render fields from a Quiqr model YAML; the drawn
  editor shows text, date, select, image, textarea, markdown, list and a
  collapsed group."

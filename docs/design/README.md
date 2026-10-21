# Quiqr ChatCMS — design hand-off (milestone 1)

UI design for the hosted Quiqr ChatCMS web app (chatcms.eu), packaged so
Claude Code can implement it in the React + Tailwind + shadcn/ui frontend.

## How to use it

1. Unzip this folder into the frontend repo, e.g. as `docs/design/`
   (it is reference material, not source code — don't import from it).
2. Copy `tokens/tokens.css` into the app's global stylesheet (or merge it with
   the `:root` block shadcn generated). Put `assets/quiqr-logo.svg` in `public/`.
3. Start Claude Code in the repo and paste the prompt from `PROMPT.md`.
   It points Claude at `DESIGN.md`, which is the actual spec.
4. Build screen by screen. After each screen, ask Claude Code to compare its
   result with the matching screenshot in `reference/screenshots/`.

Tip: if the repo has a `CLAUDE.md`, add one line to it:
`UI work follows docs/design/DESIGN.md; reference screens are in docs/design/reference/.`

## What's inside

| Path | What it is |
| --- | --- |
| `DESIGN.md` | The spec: principles, tokens, layout shell, components, every screen, behaviour, roles, copy rules. Claude Code's main input. |
| `PROMPT.md` | Ready-to-paste kick-off prompt for Claude Code, plus follow-up prompts per screen. |
| `tokens/tokens.css` | Colours, radius and fonts as shadcn-style CSS variables (Tailwind v4 `@theme inline`). |
| `assets/quiqr-logo.svg` | Official Quiqr wordmark (from quiqr.org). |
| `reference/html/` | Static HTML of every screen. Open `index.html` in a browser for a clickable overview; the screens link to each other. Exact spacing, sizes and colours are in the inline styles. |
| `reference/screenshots/desktop/` | 1440 px wide, full-page PNGs. |
| `reference/screenshots/phone/` | 390 px wide (2×) PNGs, incl. the dedicated phone screens. |
| `reference/canvas-source/` | The original Claude Design canvas files (`.dc.html` + `canvas.json`), in case you want to load them back into a canvas. |

## Status

- Milestone 1 screens: complete as static designs.
- Light theme only. Dark mode is not designed yet (tokens are structured so it can be added).
- Sample content (Korenaar Bakery, Studio Noord, people's names) is made up.
- Assistant panel, workflow editor and artifacts are placeholders for later milestones.

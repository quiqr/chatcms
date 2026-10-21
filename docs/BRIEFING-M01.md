---
status:
    - agent-draft
title: Claude Code briefing Quiqr ChatCMS milestone 1
---
# Claude Code briefing: Quiqr ChatCMS, milestone 1

Related: [[Zwakke plekken Quiqr]], [[Claude Design briefing Quiqr]]

## 0. How to work

- **Follow my 1shotpoc development workflow and its gates.** That workflow is leading: where it conflicts with this briefing on process, the workflow wins. Do not pass a gate without meeting its criteria. (1shotpoc workflow location: TODO, add path/link.)
- Spec everything with **OpenSpec** while building: one change proposal per capability, specs kept up to date with what is actually built.
- Work autonomously. Only stop for a question when a decision is truly blocking and not covered by this briefing, the screens, the reference app or the workflow. Log assumptions in the spec instead of asking.
- The UI screens from Claude Design (provided alongside this briefing) are the visual source of truth.

## 1. What we are building

Quiqr ChatCMS (chatcms.eu): a hosted, multi-tenant, Nix-powered headless CMS for Markdown-based static sites. Hugo first. Its core is generating editing forms from the declarative **Quiqr model** (YAML), so non-technical people can safely edit site content. Content lives in git; git is hidden from content creators.

Milestone 1 goal: a **fully usable MVP** I can deploy and run for a real customer whose Hugo site currently lives in the old Quiqr desktop app.

Two audiences:
- **Implementers** (template developers / site implementers): create projects, import repos, configure publishing, manage users.
- **Content creators** (customers): edit pages through forms, preview, press Publish. Never see git terminology.

## 2. References

- **Current Quiqr** (desktop app + its source) is the functional reference, especially the model language and form behaviour. Run a reference installation with the same test sites and compare.
- **Backwards compatibility**: the new app must read existing Quiqr models unchanged, and saving through a form must produce the same front matter and Markdown as the old Quiqr. Build golden tests for this.
- Leave room to extend the model language later, but do not change existing semantics.

## 3. Tech stack (locked)

- Backend: **Go**
- Frontend: **React + TypeScript + Vite + Tailwind + shadcn/ui**, fully responsive (must work on phones)
- Database: **PostgreSQL**, queries via **sqlc**, migrations via **goose** (embedded in the binary, run on startup)
- Git: call the **git CLI** from Go (not go-git), version pinned via Nix
- Packaging/deploy: **Nix flake** + NixOS module. Nix-only is fine.
- DB holds users, organisations, memberships, roles, projects, credentials, job state. Content stays in git.

## 4. Milestone 1 scope

1. **Auth**: built-in email + password login, sessions in Postgres. Design so OIDC can be added later.
2. **Tenancy & permissions**: organisations > projects > roles per project (owner, implementer, editor, viewer). All checks go through one central policy function so per-document permissions can be added later. Admin can invite users.
3. **Onboarding**: import an existing Hugo repo with a Quiqr model as a project. Quiqr generates an **SSH deploy key per project** (private key encrypted at rest); implementer adds the public key to the git host. Also support: import a site from the old desktop app, and start from a Quiqr template.
4. **Form system**: render all forms from the Quiqr model (all existing field types), save to front matter/Markdown identically to old Quiqr. Good Markdown body editor.
5. **Content management**: browse collections and singles; create, rename, delete pages.
6. **Media**: upload, browse and pick images.
7. **Git sync (must be rock solid)**: auto-commit on save, background sync jobs, detect and pull new commits from the remote, clear conflict handling, hidden from content creators. Thin internal interface so Jujutsu could replace it later.
8. **Preview**: live preview of the site (e.g. `hugo server` per editing session) showing changes within a few seconds; stop idle previews.
9. **Builds**: Hugo builds in an isolated sandbox. M1 may use a Nix sandbox; put it behind a runner interface so Firecracker micro VMs can replace it later.
10. **Workflows & actions**: declarative YAML in a `workflows` subdirectory of the site's quiqr directory. Triggers (save, preview, publish), steps that call built-in actions (hugo build, pre-publish, git push). Pre-publish step is optional/may be empty; if it fails, nothing is pushed and the user gets a friendly error.
11. **Publish**: show a plain-language overview of what changed since the last publish, then publish via git push: to main or a separate branch, pushing either the raw Hugo source (external pipeline like Cloudflare/AWS Amplify builds it) or the built HTML from Quiqr's builder.
12. **History & rollback**: list versions in plain language, show what a version contains, roll back with confirmation.
13. **Sync status indicator**: saved / syncing / remote changes pulled in.

## 5. Out of scope for M1 (but don't block it)

AI chat panel (MCP tools + direct Markdown edits; leave layout room), Quarto and artifacts, other build systems, shareable/reusable actions (Nix flakes), Firecracker runner, OIDC, GitHub App, per-document permissions, Hugo-in-WASM browser preview.

## 6. Acceptance tests (definition of done)

Run against the real customer site (and test fixtures). Automate as end-to-end tests where possible.

1. As admin, create an organisation and invite the customer as editor.
2. Import the customer's Hugo repo with its Quiqr model using a deploy key.
3. All forms from the model render; saving gives the same front matter as the old Quiqr.
4. Create, edit and delete a page, and upload an image.
5. Live preview shows changes within a few seconds.
6. A commit pushed directly to the remote shows up in Quiqr automatically.
7. Publishing shows what changed, then git push triggers the Cloudflare/Amplify deploy.
8. Roll back to an earlier version.
9. Everything works on a phone.

Plus: the app builds and deploys via the Nix flake/NixOS module, migrations run cleanly from an empty database, and all specs in OpenSpec match what was built.

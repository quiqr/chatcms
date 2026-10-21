---
# chatcms-8vem
title: 01 MVP for the first customer
status: in-progress
type: milestone
priority: critical
created_at: 2026-10-21T15:32:09Z
updated_at: 2026-10-21T15:32:09Z
---

## Goal

A fully usable MVP that can be deployed and run for a real customer whose Hugo
site currently lives in the old Quiqr desktop app.

## Definition of done

The nine acceptance tests in `docs/BRIEFING-M01.md` section 6, automated as
end-to-end tests, plus: the app builds and deploys through the Nix flake and
the NixOS module, migrations run cleanly from an empty database, and every
OpenSpec spec matches what was built.

## Hard requirement

Backwards compatibility with the old Quiqr. Existing Quiqr models must be read
unchanged, and saving through a form must produce the same front matter and
Markdown. Golden tests prove it. The model language may be extended later, but
existing semantics do not change.

## Out of scope, and not to be blocked

AI chat panel, Quarto and artifacts, other build systems, shareable actions as
Nix flakes, the Firecracker runner, OIDC, a GitHub App, per-document
permissions, and Hugo in WASM for browser preview. These live in milestone 02.

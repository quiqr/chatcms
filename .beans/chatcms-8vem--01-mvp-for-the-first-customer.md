---
# chatcms-8vem
title: 01 MVP for the first customer
status: in-progress
type: milestone
priority: critical
created_at: 2026-10-21T15:32:09Z
updated_at: 2026-10-26T16:45:04Z
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

## Summary of Changes

Milestone 01 is done. Every epic under it is closed, and the nine acceptance
tests in `docs/BRIEFING-M01.md` section 6 pass by themselves in
`nix flake check`, in a real browser, against the binary this flake builds, at
1440 px and at 390 px.

The three conditions this milestone added to those nine also hold: the app
builds and deploys through the flake and the NixOS module, proven by a virtual
machine test that boots it; migrations run from an empty database on every
integration run; and every capability a change declared has a spec, which
`scripts/check-specs.sh` enforces before anything ships.

### The epics

| Epic | What it added |
| ---- | -------------- |
| Project foundation and the gate | The flake, the gate, the server, the schema |
| Quiqr model parser | Reading the desktop app's models unchanged, all twenty-seven field types |
| Authentication and sessions | Signing in, invitations, sessions |
| Tenancy, roles and the central policy | One place that answers who may do what |
| Design foundation | Tokens, components, shells, at both widths |
| Git sync engine | The git CLI behind one package, versions, exchange with a remote |
| Form rendering and round-trip save | A form from the model, and front matter written back byte for byte |
| Project onboarding and deploy keys | Bringing a site in with a key of its own |
| Sandboxed build runner | Builds with a time and memory limit and no network |
| Declarative workflows | What a site says happens when it is saved or published |
| Publish flow | A review of what changed, then the publish |
| History and rollback | A timeline in plain language, and going back without losing anything |
| Live preview | The site as a reader sees it, changing as you type |
| Media library | Pictures, descriptions, and where each one is used |
| Content management | Making, renaming, copying and removing pages |
| Sync status and conflict surfacing | Five states, and a question per field when two people collide |
| Deployment | The NixOS module, a virtual machine test, and how to release |
| End-to-end acceptance suite | The nine acceptance tests, automated, in the gate |

### What is deliberately not here

The assistant panel, Quarto, other build systems, shareable actions, the
Firecracker runner, OIDC, a GitHub App, per-document permissions and Hugo in
WASM. They are milestone 02, and nothing in milestone 01 was shaped to avoid
them.

### Known gaps, written down rather than quietly dropped

`docs/design/COMPARISON.md` lists where the screens still differ from the
drawings they came from, and which of those differences were decided on
purpose. The ones still to do are small and none of them is an acceptance
test.

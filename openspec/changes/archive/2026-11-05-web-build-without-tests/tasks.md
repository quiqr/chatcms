# Tasks

## 1. Build

- [x] 1.1 Give `mkWeb` a `checked` flag that sets `doCheck`, off by default, and pass `checked = true` for `checks.frontend`; verify `nix build .#web` runs no Vitest and `nix build .#checks.x86_64-linux.frontend` still does
- [x] 1.2 Add an entry under `[Unreleased]` in `CHANGELOG.md`

## 2. Integration

- [x] 2.1 Run `nix flake check` and verify it is green

## Workflow follow-up

- Ship with `scripts/ship-change.sh web-build-without-tests "<subject>"`, then bump `chatcms` in mipnix and deploy durer.

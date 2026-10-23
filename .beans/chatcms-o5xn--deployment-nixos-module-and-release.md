---
# chatcms-o5xn
title: 'Deployment: NixOS module and release'
status: in-progress
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T15:38:35Z
parent: chatcms-8vem
blocked_by:
    - chatcms-io95
---

## Purpose

Deploy it. The flake and the NixOS module are the only supported path, and both
must be proven, not assumed.

## Scope

- Finish `nix/nixos-module.nix`: options for the database, the secret key file,
  the state directory, the base URL, and the pinned git and Hugo binaries.
- A NixOS VM test that boots the module, waits for the service, runs the
  migrations from an empty database and fetches the sign-in page.
- The `nixos-vm` check in `flake.nix`.
- Deployment notes: how to provide the secret key file, how to back up Postgres
  and the state directory, and what to do on a failed migration.
- Release process: version stamping, the CHANGELOG promotion step, and tagging.

## Done when

- A fresh NixOS VM runs the app from the module with nothing but the options
  documented.
- Migrations run cleanly from an empty database in the VM test.
- The secret key file is never in the Nix store, asserted by the test.
- `nix flake check --all-systems` evaluates for every supported system.

## References

- `docs/BRIEFING-M01.md` section 6, the closing paragraph.
- `nix/nixos-module.nix`.

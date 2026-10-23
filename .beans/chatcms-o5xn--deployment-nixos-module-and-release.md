---
# chatcms-o5xn
title: 'Deployment: NixOS module and release'
status: completed
type: epic
priority: normal
created_at: 2026-10-21T15:32:21Z
updated_at: 2026-10-23T22:00:00Z
openspec-link: openspec/changes/archive/2026-10-23-add-deployment
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-10-23-add-deployment`.

A NixOS virtual machine now boots the module in the gate: the database and the
service come up, the migrations run from an empty database, the sign-in page
is served, the service restarts with its data intact, and the pinned git, Hugo
and sandbox are the ones it uses.

**The test failed on its first run exactly as a first deployment would have:**
a master key owned by root with mode 0600, and a service running as
`chatcms`. The obvious fix is to widen the key's permissions, which is the
wrong direction for a secret. systemd's `LoadCredential` reads it as root and
hands the service a copy only that service can see, so the key stays
root-owned and the service stays non-root. The test asserts all three: still
`root 600`, the service started, and no copy of its contents anywhere in the
Nix store.

Absence from the store is searched for rather than assumed: a path outside the
store proves nothing if something copied the bytes in.

`docs/DEPLOYING.md` and `docs/RELEASING.md` are written, and the changelog
says what this milestone added.

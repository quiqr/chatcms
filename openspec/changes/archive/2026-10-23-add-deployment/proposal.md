## Why

"It deploys" has been a thing somebody would have had to try by hand. Epic
`chatcms-o5xn`, and the part of the definition of done that says the app
deploys via the flake and the NixOS module with migrations running cleanly
from an empty database.

## What Changes

- **A NixOS virtual machine test** that boots the module, waits for the
  service, checks the migrations ran from an empty database, fetches the
  sign-in page, restarts the service and checks it comes back.
- **The master key is handed over by systemd**, so it stays owned by root and
  the service does not run as root to read it. The test asserts that it stays
  root-owned, that no copy of its contents is in the Nix store, and that the
  service starts anyway.
- **`nixos-vm` is part of the gate** on Linux, so a change that breaks
  deployment fails before it is shipped rather than when somebody deploys.
- **`docs/DEPLOYING.md`**: the smallest configuration that works, what to back
  up, what to do when a migration fails, and why builds may be unavailable.
- **`docs/RELEASING.md`**: version stamping, the changelog promotion and the
  tag.

## Capabilities

### New Capabilities

- `platform/deployment`: what a machine running ChatCMS is, and what it
  promises about the master key and the database.

### Modified Capabilities

None.

## Impact

- **New code**: `nix/vm-test.nix`, the `nixos-vm` check.
- **Changed**: the module hands the master key to the service through systemd
  rather than expecting the file to be readable by the service user.
- **New docs**: `docs/DEPLOYING.md`, `docs/RELEASING.md`, and the changelog.

## Assumptions

- **The service never reads the master key file itself.** systemd reads it as
  root and gives the service a copy only that service can see. The first run
  of the test failed exactly as a first deployment would have: a root-owned
  key and a non-root service. Making the file readable by the service user
  would have worked and would have been worse, because the obvious way to do
  that is to widen the permissions on a secret.

- **The virtual machine test runs on Linux only.** A machine cannot be booted
  on a Mac, so the check exists there and not elsewhere; everything else still
  runs on every system.

- **The test asserts the absence of the key from the store, not merely that
  the path is outside it.** A path outside the store proves nothing if
  something copied the contents in. It searches the store for the key's
  contents.

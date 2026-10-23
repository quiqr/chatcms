# Tasks

## 1. The machine test

- [x] 1.1 Boot the module, wait for the database and the service
- [x] 1.2 Check the migrations ran from an empty database
- [x] 1.3 Fetch the sign-in page
- [x] 1.4 Restart and check it comes back with the database as it was
- [x] 1.5 Check the service runs as its own user
- [x] 1.6 Check the pinned git, Hugo and sandbox are the ones used

## 2. The master key

- [x] 2.1 Hand it to the service through systemd rather than widening it
- [x] 2.2 Check it stays owned by root
- [x] 2.3 Search the store for its contents and find nothing

## 3. The gate

- [x] 3.1 `nixos-vm` as a check, on Linux
- [x] 3.2 `nix flake check` green with it

## 4. The words

- [x] 4.1 `docs/DEPLOYING.md`
- [x] 4.2 `docs/RELEASING.md`
- [x] 4.3 The changelog says what this milestone added

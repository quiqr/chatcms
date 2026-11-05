# Tasks

## 1. Seed

- [x] 1.1 Give `marieke@korenaar.example` and `sam@korenaar.example` the seeded passphrase in `cmd/seed`, and make sure their membership no longer reads as still to accept; verify with an integration test that seeds a database and signs in as the editor and as the viewer, each seeing `korenaar` with their role
- [x] 1.3 Keep one invitation open in the seed (`noor@korenaar.example`, editor) and point acceptance test 19 at her; verify with the seed integration test and the `e2e` check
- [x] 1.2 List the editor and viewer in the seed's output and in the notice of `scripts/dev.sh`; verify by running `CHATCMS_INSTANCE=test bash scripts/dev.sh fresh` and reading the notice

## 2. Development sign-in block

- [x] 2.1 Add the labels for the block and its four buttons to `web/src/strings/en.ts` and every other catalogue; verify with `npm run typecheck`
- [x] 2.2 Add `DevSignIn.tsx` with the four people and the passphrase, loaded by `SignInPage.tsx` only inside `import.meta.env.DEV`; each button submits the ordinary sign-in; verify with a Vitest test that pressing the editor button posts the editor's address and passphrase to the sign-in endpoint
- [x] 2.3 Verify with a Vitest test that the block is absent when `import.meta.env.DEV` is false
- [x] 2.4 Verify `npm run check:strings` and `npm run check:interface` pass, and that the block works at 390 px

## 3. Production bundle

- [x] 3.1 Add a script that fails when the built frontend contains the seeded passphrase or addresses, and run it from the `frontend` check in `flake.nix`; verify it fails on a bundle with the passphrase planted and passes on the real build

## 4. Documentation

- [x] 4.1 Mention the buttons in the comment at the top of `scripts/dev.sh` and in the dev section of `CLAUDE.md`; verify by reading both
- [x] 4.2 Add an entry under `[Unreleased]` in `CHANGELOG.md`

## 5. Integration

- [x] 5.1 Run `nix flake check` and verify it is green

## Workflow follow-up

- Ship with `scripts/ship-change.sh dev-sign-in-as-role "<subject>"`.

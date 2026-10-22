## 1. Secrets

- [x] 1.1 Add `internal/secrets` sealing and opening values with AES-256-GCM
  under a master key, with a fresh nonce each time; verify with
  `TestSealedValueIsNotTheSecret` and `TestOpeningReturnsWhatWasSealed`.
- [x] 1.2 Make the same secret seal differently each time; verify with
  `TestSealingTwiceDiffers`.
- [x] 1.3 Fail to open a value that was altered; verify with
  `TestTamperedValueWillNotOpen` for a changed nonce and a changed ciphertext.
- [x] 1.4 Read the master key from the file configuration names, refusing a
  missing or too-short key at start-up; verify with
  `TestMasterKeyIsReadFromItsFile`, `TestMissingKeyRefusesToStart` and
  `TestShortKeyRefusesToStart`.
- [x] 1.5 Add `CHATCMS_SECRET_KEY_FILE` to the configuration and require it;
  verify with `TestLoadConfigRequiresTheSecretKeyFile`.

## 2. Keys

- [x] 2.1 Generate an Ed25519 pair and marshal both halves in the OpenSSH
  formats, without running any other program; verify with
  `TestGeneratedKeyIsUsableByGit` by handing the public half to a local server
  and the private half to git.
- [x] 2.2 Store the private half sealed and the public half as it is; verify
  with `TestPrivateHalfIsStoredSealed`.
- [x] 2.3 Give each project its own pair; verify with `TestKeysDifferPerProject`.
- [x] 2.4 Replace a project's pair on request; verify with `TestReplacingAKey`.
- [x] 2.5 Never return or log the private half; verify with
  `TestPrivateKeyIsNeverReturned` over every endpoint that mentions a project,
  and `TestPrivateKeyIsNeverLogged`.

## 3. Reaching a remote with the project's key

- [x] 3.1 Make a directory for decrypted keys at start-up with owner-only
  permission, and remove it on shutdown; verify with
  `TestKeyDirectoryIsPrivate` and `TestNoKeyRemainsAfterShutdown`.
- [x] 3.2 Write a project's key there when it is first needed, owner-only;
  verify with `TestKeyFileIsOwnerOnly`.
- [x] 3.3 Hand `internal/gitsync` the project's key rather than a shared one;
  verify with `TestEachProjectUsesItsOwnKey`.
- [x] 3.4 Say that a project has no key yet rather than reporting a refusal
  from the remote; verify with `TestNoKeyYetIsItsOwnMessage`.

## 4. Importing

- [x] 4.1 Import from a git repository; verify with `TestImportFromARepository`
  against a bare repository on disk.
- [x] 4.2 Import from a directory holding an exported site, making it a
  repository and sending it to the remote; verify with `TestImportFromAnExport`.
- [x] 4.3 Import from a template, which is a repository ChatCMS knows; verify
  with `TestImportFromATemplate`.
- [x] 4.4 Give the same project for an export and the repository made from it;
  verify with `TestAnExportAndItsRepositoryAgree`.
- [x] 4.5 Record where the project came from; verify with
  `TestProjectRecordsItsSource` and that a project with nothing imported says
  so.

## 5. Inspecting what arrived

- [x] 5.1 Detect a Hugo site by its configuration, under the names and places
  Hugo accepts; verify with `TestDetectsHugoConfiguration` over each.
- [x] 5.2 Detect a Quiqr model through `internal/model`; verify with
  `TestDetectsAModel`.
- [x] 5.3 Report a Hugo site with no model as usable and needing one; verify
  with `TestHugoWithNoModelIsUsable`.
- [x] 5.4 Report something that is not a Hugo site, naming what was looked
  for; verify with `TestNotAHugoSite`.
- [x] 5.5 Report a model that does not parse, with its own error; verify with
  `TestAModelThatDoesNotParse`.

## 6. Testing the connection

- [x] 6.1 Report that the remote took the key; verify with
  `TestConnectionIsGood` against a local repository.
- [x] 6.2 Report a refused key with what to do next and the key again; verify
  with `TestKeyNotAddedYet`.
- [x] 6.3 Report an unreachable remote, distinguished from a refusal; verify
  with `TestRemoteUnreachable`.
- [x] 6.4 Carry no secret in any of these answers; verify with
  `TestConnectionAnswersCarryNoSecret`.

## 7. Publishing settings

- [x] 7.1 Record publishing to the main line or to one of its own, with a
  default name offered; verify with `TestPublishingTarget`.
- [x] 7.2 Refuse a name a remote cannot accept, saying what is allowed; verify
  with `TestRefusesAnUnusableBranchName`.
- [x] 7.3 Record publishing the source or the built result; verify with
  `TestPublishingForm`.
- [x] 7.4 Allow the settings only to those who may change settings; verify
  with `TestOnlyThoseWhoMayChangePublishing`.

## 8. The screen

- [x] 8.1 Build the onboarding screen with its step list and three sections as
  the designs draw it; verify with `shows the three steps` and compare with
  `docs/design/reference/screenshots/desktop/onboarding.png`.
- [x] 8.2 Show the public key with a way to copy it, and never the private
  one; verify with `shows the key to copy`.
- [x] 8.3 Offer the publishing choices as the designs draw them; verify with
  `offers where and what to publish`.
- [x] 8.4 Check it at 390 pixels in a browser and record the result.

## 9. Acceptance and shipping

- [x] 9.1 Prove acceptance test 2 end to end: set up a project, show its key,
  add it to a remote, import the site, and see the model found; verify with
  `TestAcceptanceImportWithADeployKey`.
- [x] 9.2 Confirm the coverage gate passes with the new packages.
- [x] 9.3 Add bullets to `CHANGELOG.md` under `## [Unreleased]`.
- [x] 9.4 Run `nix flake check` and confirm every check passes.

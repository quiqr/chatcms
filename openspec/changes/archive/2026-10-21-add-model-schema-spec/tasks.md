## 1. The missing spec

- [x] 1.1 Write `specs/content/model-schema/spec.md` from the tests in
  `internal/model`, asserting nothing the tests do not prove; verify each
  scenario against its test in task 1.2.
- [x] 1.2 Match every scenario to its test: `TestRootIsTyped`,
  `TestHugoverIsMigratedAndPreserved`, `TestExplicitGeneratorWinsOverHugover`,
  `TestEveryFieldTypeIsTyped`, `TestSelectOptionsAcceptBothShapes`,
  `TestContainersNestToAnyDepth`, `TestLeafArrayCarriesOneChildField`,
  `TestUnknownFieldTypeSurvives`, `TestUnknownKeysArePreserved`,
  `TestErrorNamesFileAndPath`, `TestWrongTypeForAKnownKeyIsRejected` and
  `TestAListKeyThatIsNotAListIsAWarning`; verify by running
  `go test ./internal/model/ -run` for each name and seeing it pass.

## 2. Stop it happening again

- [x] 2.1 Add `scripts/check-specs.sh`, which reads the capability paths a
  change's `proposal.md` declares and fails when any of them has no
  `specs/<path>/spec.md` in that change; verify it fails on a change with a
  declared capability whose spec file is removed, and passes on this one.
- [x] 2.2 Call it from `scripts/ship-change.sh` before the gate, so a change
  cannot be archived with a capability it never specified; verify by running
  the script against a deliberately incomplete change.
- [x] 2.3 Note the check in `AGENTS.md` under the gate; verify the command in
  it runs as written.

## 3. Shipping

- [x] 3.1 Confirm `openspec validate add-model-schema-spec` passes and the
  change carries no code change.
- [x] 3.2 Note in `CHANGELOG.md` only if there is a user-facing effect; there is
  none, so record in the ship report that the CHANGELOG was deliberately
  untouched.
- [x] 3.3 Run `nix flake check` and confirm every check passes.

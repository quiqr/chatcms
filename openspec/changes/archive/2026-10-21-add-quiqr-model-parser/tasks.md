## 1. The inventory, written first

- [x] 1.1 Write `docs/QUIQR-MODEL.md` from the reference app at
  `../quiqr-desktop`: the file layout, the assembly order, the merge rules, the
  root properties, and a table per field type with its properties, defaults and
  meaning; verify every type in the reference app's `FieldRegistry.ts` appears.
- [x] 1.2 Record in the same document the two merge directions and the field
  deduplication rule, with the reference app file and line each came from;
  verify by re-reading those lines.

## 2. Format providers

- [x] 2.1 Add `internal/model`, with a format provider that turns bytes plus a
  file extension into a generic tree, for YAML, TOML and JSON; verify with
  `TestProviderParsesEachFormat` that the same document in all three formats
  yields equal trees.
- [x] 2.2 Decode numbers without widening them: integers stay integers in every
  format; verify with `TestNumbersSurviveEachFormat`.
- [x] 2.3 Return a parse error naming the file and the line and column; verify
  with `TestParseErrorNamesFileAndPosition` for each format.
- [x] 2.4 Reject a model file whose top level is not a mapping; verify with
  `TestTopLevelMustBeAMapping`.

## 3. The assembler

- [x] 3.1 Implement the merge primitive with the reference semantics: maps
  recursively, arrays concatenated target-first, anything else source-wins;
  verify with a table test `TestMergeSemantics` covering each row.
- [x] 3.2 Locate the base file: `quiqr/model/base.*`, then `sukoh.*` at the
  root, then "no model"; verify with `TestLocateBaseFile` over an `fstest.MapFS`
  for each case, and that "no model" is distinguishable from a parse failure.
- [x] 3.3 Merge `quiqr/model/includes/*` in filename order, appending to the
  list keys and deep-merging everything else with the base winning; verify with
  `TestIncludesAppendToLists`, `TestIncludesMergeObjectsUnderTheirFilename` and
  `TestBaseWinsOverInclude`.
- [x] 3.4 Append each file under `includes/singles/`, `includes/collections/`
  and `includes/menus/` as one item; verify with `TestSubIncludesAddOneItemEach`
  and `TestBothIncludeStylesAtOnce`.
- [x] 3.5 Resolve `_mergePartial` from `quiqr/model/partials/<name>.*` and from
  `file://`, merge with the item winning, dedupe fields keeping the last
  occurrence, and drop the `_mergePartial` key; verify with
  `TestPartialSuppliesFields`, `TestItemOverridesPartial` and
  `TestPartialFieldOrderMatchesReference`.
- [x] 3.6 Resolve an `http://` or `https://` partial from
  `quiqr/model/partialsRemoteCache` only, and error naming the partial when the
  cache has no copy; verify with `TestRemotePartialFromCache` and
  `TestUncachedRemotePartialIsAnError`.
- [x] 3.7 Report a `dogfood_site://` partial as a model warning naming it,
  rather than failing; verify with `TestDogfoodPartialIsAWarning`.
- [x] 3.8 Make assembly read-only and deterministic; verify with
  `TestAssemblyWritesNothing` using a read-only FS and
  `TestAssemblyIsDeterministic` comparing two runs.

## 4. The schema

- [x] 4.1 Type the root: singles, collections, dynamics, menu, build, serve,
  generator type and version; verify with `TestRootIsTyped`.
- [x] 4.2 Migrate `hugover` to the generator type and version while preserving
  the original key; verify with `TestHugoverIsMigratedAndPreserved`.
- [x] 4.3 Type every field in the inventory, each with its own configuration
  struct where it has properties of its own; verify with a table test
  `TestEveryFieldTypeIsTyped` driven by the inventory, which fails when a type
  is missing.
- [x] 4.4 Carry container children at any depth, including `leaf-array`'s single
  child field; verify with `TestContainersNestToAnyDepth`.
- [x] 4.5 Keep an unrecognised field type as a placeholder carrying its declared
  type and raw properties; verify with `TestUnknownFieldTypeSurvives`.
- [x] 4.6 Preserve unrecognised keys on the model, on singles and collections,
  and on fields; verify with `TestUnknownKeysArePreserved`.
- [x] 4.7 Produce errors carrying file, path and expectation; verify with
  `TestErrorNamesFileAndPath` for a collection without a folder, a nested field
  without a key, and two singles sharing a key.

## 5. Content navigation

- [x] 5.1 Derive the navigation from the model, with pages, collections and
  per-collection item counts read through the same `fs.FS`; verify with
  `TestNavigationListsPagesAndCollections` and `TestCollectionCountsItems`.
- [x] 5.2 Count an absent or empty collection folder as none rather than
  failing; verify with `TestEmptyCollectionStillAppears`.
- [x] 5.3 Carry each collection's singular item name, with a neutral fallback
  that uses no git or developer vocabulary; verify with
  `TestCollectionCarriesItemName` and `TestItemNameFallbackIsPlain`.
- [x] 5.4 Follow the model's menu when it declares one, and otherwise put
  singles before collections in declaration order; verify with
  `TestMenuDecidesOrder` and `TestWithoutMenuSinglesComeFirst`.
- [x] 5.5 Omit a menu entry naming something the model does not declare, and
  report it as a warning; verify with `TestMenuEntryWithoutATargetIsAWarning`.
- [x] 5.6 Hide entries the model marks hidden while keeping them reachable by
  key; verify with `TestHiddenCollectionIsNotInNavigation`.

## 6. Fixtures and the JSON shape

- [x] 6.1 Build `test/fixtures/models/` with a fixture per concern: every field
  type, all three formats, includes, sub-includes, partials, a legacy `sukoh`
  model, and one taken from the reference app's own test fixtures; verify each
  assembles without error.
- [x] 6.2 Pin the JSON representation with golden files per fixture; verify with
  `TestModelJSONMatchesGolden`, and that changing a field's shape fails it.
- [x] 6.3 Add a regeneration path for the goldens that is explicit, documented
  in `docs/TESTING.md`, and never runs during a normal test run; verify the
  goldens are unchanged after running it.

## 7. The gate and shipping

- [x] 7.1 Add the new dependencies and update `vendorHash` in `flake.nix`;
  verify `nix build .#chatcmsd` succeeds.
- [x] 7.2 Confirm `internal/model` is at or above the 80% core-package floor;
  verify the gate no longer reports it as absent.
- [x] 7.3 Add a bullet to `CHANGELOG.md` under `## [Unreleased]` and
  `### Added`, written from a site implementer's perspective.
- [x] 7.4 Run `nix flake check` and confirm every check passes.

## 1. Goldens, generated before anything is written

- [x] 1.1 Add `scripts/generate-goldens.sh`, running Node with `js-yaml` and
  `tomlify-j0.4` over a table of values and writing the result under
  `test/fixtures/content/golden/`; verify it produces output and records the
  versions it used.
- [x] 1.2 Cover the scalar table: a plain string, one that reads as a number,
  one that reads as a boolean, one with an apostrophe, one with a colon, one
  starting with a hash or a dash, an empty string, and null; verify each
  appears in the generated goldens.
- [x] 1.3 Cover multi-line strings, a string just under the fold width, one
  just over, one far over, and one long string with no space to break at;
  verify each appears in the generated goldens.
- [x] 1.4 Cover lists of scalars, lists of objects, nested objects, an empty
  list and an empty object; verify each appears.
- [x] 1.5 Cover numbers and dates: integers, negatives, floats, a large
  integer, and a date; verify each appears.
- [x] 1.6 Document the script in `docs/TESTING.md` alongside the model
  goldens, saying it is run by hand and never by the gate; verify the command
  in it runs as written.

## 2. The YAML emitter

- [x] 2.1 Add `internal/frontmatter` with a YAML emitter producing two-space
  indentation and sequences indented under their key; verify with
  `TestYAMLMatchesGoldenScalars` reading the generated goldens.
- [x] 2.2 Implement the quoting decision: bare when unambiguous, single quotes
  otherwise, with the reference's escaping inside them; verify with
  `TestYAMLQuotingMatchesGolden` over the scalar table.
- [x] 2.3 Implement literal blocks for multi-line strings, including the
  chomping indicator the reference chooses; verify with
  `TestYAMLMultiLineMatchesGolden`.
- [x] 2.4 Implement folding at the reference's line width, including the case
  with no break opportunity; verify with `TestYAMLFoldingMatchesGolden` and
  `TestFoldedStringReadsBackUnchanged`.
- [x] 2.5 Implement null, empty string, empty list and empty object, and keep
  the four distinguishable; verify with `TestYAMLEmptyThingsMatchGolden`.
- [x] 2.6 Implement lists, lists of objects and nested objects at any depth;
  verify with `TestYAMLCollectionsMatchGolden`.
- [x] 2.7 Implement numbers and dates as the reference spells them; verify
  with `TestYAMLNumbersAndDatesMatchGolden`.
- [x] 2.8 Refuse a shape the emitter has not been proven against, naming the
  path and the kind; verify with `TestEmitterRefusesAnUnsupportedShape`.

## 3. TOML and JSON

- [x] 3.1 Record the evidence that the reference writes TOML it cannot read
  back, as a test that runs the reference's own reader regular expression over
  the reference's own writer output; verify with
  `TestReferenceTOMLOutputIsUnreadableByItsOwnReader`.
- [x] 3.2 Read TOML and JSON front matter into values; verify with
  `TestReadsTOMLFrontMatter` and `TestReadsJSONFrontMatter`.
- [x] 3.3 Refuse to write TOML or JSON front matter, with a message naming the
  format; verify with `TestWritingTOMLIsRefused` and
  `TestWritingJSONIsRefused`, and that the message names the format.

## 4. The document

- [x] 4.1 Split a content file into front matter and body, choosing the format
  by the opening delimiter, for all three formats and for a file with none;
  verify with `TestSplitsEachFormat` and `TestFileWithNoFrontMatter`.
- [x] 4.2 Handle a file whose front matter closes with nothing after it, and a
  body that contains a delimiter-looking line; verify with
  `TestFrontMatterWithNoBody` and `TestBodyContainingADelimiter`.
- [x] 4.3 Fail on front matter that does not parse, naming the file and the
  position, returning nothing; verify with `TestUnparseableFrontMatter`.
- [x] 4.4 Keep the key order the file had, and append new keys in the order
  the model declares them; verify with `TestKeyOrderIsKept` and
  `TestNewKeysFollowTheModel`.
- [x] 4.5 Write a document back, reproducing the delimiters and the blank line
  the reference puts between them and the body; verify with
  `TestWrittenLayoutMatchesGolden`.

## 5. Values and fields

- [x] 5.1 Add `internal/forms` mapping a document onto a model's fields and
  back; verify with `TestPlainFieldRoundTrips`.
- [x] 5.2 Implement `section` and `nest` nesting, including `groupdata: false`;
  verify with `TestSectionNests` and `TestSectionWithoutGroupdataDoesNot`.
- [x] 5.3 Implement `pull` nesting under its group or its key; verify with
  `TestPullUsesItsGroup` and `TestPullWithoutAGroupUsesItsKey`.
- [x] 5.4 Implement `accordion` as a list of objects and `leaf-array` as a
  list of the child's values; verify with `TestAccordionIsAListOfObjects` and
  `TestLeafArrayIsAListOfValues`.
- [x] 5.5 Support grouping inside grouping at any depth; verify with
  `TestNestingToAnyDepth`.
- [x] 5.6 Read and write the body through the field whose key names it; verify
  with `TestBodyIsReadableAsAField` and `TestWritingTheBodyLeavesValuesAlone`.
- [x] 5.7 Preserve values no field claims, in place; verify with
  `TestUnclaimedValuesSurviveARoundTrip` and
  `TestUnclaimedValuesSurviveAnEdit`.

## 6. The contract

- [x] 6.1 Build `test/fixtures/content/` with documents in each format,
  covering every field type the model parser knows, produced through the
  reference implementation; verify each parses.
- [x] 6.2 Assert a round trip with no edit is byte identical for every
  fixture; verify with `TestRoundTripIsByteIdentical`, including the final
  newline.
- [x] 6.3 Assert an edit per field type matches its golden; verify with
  `TestEditedDocumentMatchesGolden`.
- [x] 6.4 Assert an edit changes one thing: compare the before and after bytes
  and require the difference to be confined to the edited value; verify with
  `TestAnEditChangesOnlyWhatWasEdited`.
- [x] 6.5 Assert a cleared value is written as the reference writes an absent
  value; verify with `TestClearedValueMatchesGolden`.

## 7. The gate and shipping

- [x] 7.1 Confirm `internal/forms` is at or above the 80% core-package floor
  and `internal/frontmatter` is well covered; verify the gate no longer
  reports forms as absent.
- [x] 7.2 Add a bullet to `CHANGELOG.md` under `## [Unreleased]`, written from
  a site implementer's perspective.
- [x] 7.3 Record in `docs/TESTING.md` which versions of the reference packages
  produced the goldens; verify the versions in it match the generator's
  record.
- [x] 7.4 Run `nix flake check` and confirm every check passes.

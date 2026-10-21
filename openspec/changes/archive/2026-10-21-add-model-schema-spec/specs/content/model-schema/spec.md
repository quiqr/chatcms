## Purpose

Defines what an assembled Quiqr model contains, which field types exist, and
what the system does with parts of a model it does not recognise, so that a
model written for any version of Quiqr can still be opened.

## ADDED Requirements

### Requirement: A model describes singles, collections and a menu

An assembled model SHALL expose the site's singles, its collections, its
dynamics, its menu, its build and serve entries, and its static site generator
type and version.

#### Scenario: A single names its file

- **WHEN** a model declares a single
- **THEN** the single carries its key, its title, the content file it edits and
  its fields

#### Scenario: A collection names its folder

- **WHEN** a model declares a collection
- **THEN** the collection carries its key, its title, the folder it lists, the
  file extension and data format of its items, the key used as an item title,
  and its fields

#### Scenario: The legacy generator version key is migrated

- **WHEN** a model declares `hugover` and neither `ssgType` nor `ssgVersion`
- **THEN** the assembled model reports the generator type as Hugo and the
  version from `hugover`
- **AND** the original key is preserved so writing the model back does not lose
  it

#### Scenario: An explicit generator version wins

- **WHEN** a model declares both `hugover` and `ssgVersion`
- **THEN** the assembled model reports the explicit `ssgVersion`

#### Scenario: The generator defaults to Hugo

- **WHEN** a model declares neither a generator type nor `hugover`
- **THEN** the assembled model reports Hugo

### Requirement: Every field type the desktop app supports is recognised

The system SHALL recognise each field type of the reference application, with
its own properties: the data fields `string`, `markdown`, `easymde`,
`readonly`, `uniq`, `number`, `slider`, `boolean`, `date`, `select`,
`select-from-query`, `chips`, `color`, `font-picker`, `fonticon-picker`,
`image-select` and `eisenhouwer`; the container fields `section`, `nest`,
`accordion`, `pull`, `leaf-array` and `bundle-manager`; and the layout fields
`hidden`, `empty-line`, `info` and `bundle-image-thumbnail`.

#### Scenario: Every declared property is consumed

- **WHEN** a field of any recognised type declares the properties that type
  accepts
- **THEN** the assembled field carries them, typed
- **AND** none of them is left over as an unrecognised key

#### Scenario: Properties shared by every field

- **WHEN** any field declares a key, title, tip, default, hidden, group,
  groupdata, name, content, theme, array title or auto-save
- **THEN** the assembled field carries them regardless of its type

#### Scenario: A select accepts both option shapes

- **WHEN** a `select` declares options as bare values and as mappings with a
  value and a display text
- **THEN** both shapes are carried, with the value and the text distinguished

#### Scenario: A container field carries its children

- **WHEN** a model declares an `accordion` whose fields include a `section`
  whose fields include a `nest`
- **THEN** the assembled model carries the whole tree, at any depth

#### Scenario: A leaf array carries one child field

- **WHEN** a model declares a `leaf-array`
- **THEN** the assembled field carries the single field every item of the array
  has, and no list of children

### Requirement: An unknown field type survives as a placeholder

A field whose type the system does not recognise SHALL be kept as a placeholder
carrying its key, its declared type and its raw properties. It SHALL NOT be
dropped and SHALL NOT fail the model.

#### Scenario: A model from a newer Quiqr

- **WHEN** a model declares a field of a type this version does not know
- **THEN** the model loads
- **AND** the field is present, marked as an unrecognised type, naming the type
  it declared and carrying the properties it declared

### Requirement: Unknown keys are preserved

Keys the system does not recognise SHALL be preserved on the model, on each
single and collection, and on each field, so that reading a model and writing
it back loses nothing.

#### Scenario: A round trip keeps an unknown key

- **WHEN** a model declares a key this version has no meaning for, at the root,
  on a single, on a collection and on a field
- **AND** the model is read
- **THEN** each of those keys and its value is still carried

### Requirement: An invalid model says what is wrong and where

When a model cannot be understood, the error SHALL name the file, the path
within the model, and what was expected.

#### Scenario: A collection without a folder

- **WHEN** a collection declares no folder
- **THEN** the error names the file and the path to the missing property

#### Scenario: A field without a key

- **WHEN** a field inside a nested container declares no key
- **THEN** the error names the file and the full path to that field

#### Scenario: A required property of a field type is missing

- **WHEN** a `slider` declares no maximum, or a `select` declares no options
- **THEN** the error names the file and the path to the missing property

#### Scenario: Two singles with the same key

- **WHEN** two singles, or two collections, declare the same key
- **THEN** the error names the duplicated key

#### Scenario: A known key of the wrong kind

- **WHEN** a key that must be a string holds a number, or a key that must be a
  list holds something else
- **THEN** the model is rejected rather than coerced

### Requirement: A list key that is not a list is reported, not silently dropped

The reference application replaces `singles`, `collections`, `dynamics` or
`menu` with an empty list when what it finds there is not a list. The system
SHALL do the same so that assembly matches, and SHALL report what it ignored.

#### Scenario: Singles declared as a mapping

- **WHEN** a model declares `singles` as a mapping rather than a list
- **THEN** the model loads with no singles
- **AND** a warning names `singles` and says what was there was ignored

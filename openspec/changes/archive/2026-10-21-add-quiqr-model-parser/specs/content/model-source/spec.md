## Purpose

Defines how a site's Quiqr model is located and assembled from the files that
make it up, so that the same site read by ChatCMS and by the Quiqr desktop app
yields the same model.

## ADDED Requirements

### Requirement: The model is found where Quiqr puts it

The system SHALL look for a site's model base file at `quiqr/model/base` with
any supported extension, and SHALL fall back to `sukoh` with any supported
extension at the site root. A site with neither SHALL be reported as having no
Quiqr model, not as an error in the model.

#### Scenario: A model in the current location

- **WHEN** a site contains `quiqr/model/base.yaml`
- **THEN** that file is the base of the assembled model

#### Scenario: A model in the legacy location

- **WHEN** a site contains no `quiqr/model/base.*` but does contain `sukoh.yaml`
- **THEN** that file is the base of the assembled model

#### Scenario: No model at all

- **WHEN** a site contains neither
- **THEN** the result says the site has no Quiqr model
- **AND** it is distinguishable from a model that failed to parse

### Requirement: Three formats are read, chosen by extension

Model files SHALL be read as YAML, TOML or JSON according to their extension,
and a single site SHALL be able to mix formats across its model files.

#### Scenario: A TOML base with YAML includes

- **WHEN** a site has `quiqr/model/base.toml` and `quiqr/model/includes/collections.yaml`
- **THEN** both are read with the right parser
- **AND** the assembled model contains the collections from the include

#### Scenario: A file that is not valid in its format

- **WHEN** a model file cannot be parsed
- **THEN** the error names that file and the position of the problem
- **AND** no partially assembled model is returned

### Requirement: Includes are merged by filename

Files directly under `quiqr/model/includes/` SHALL be merged into the model
under the key given by their filename without extension. A file holding an
array whose name is one of the model's list keys SHALL be appended to that
list; any other file SHALL be deep-merged as an object, with values already in
the base file winning over the include.

#### Scenario: An array include extends a list

- **WHEN** `includes/collections.yaml` holds an array of collections
- **AND** the base file already declares collections
- **THEN** the assembled model has both, with the base file's entries first

#### Scenario: An object include is merged under its filename

- **WHEN** `includes/menu.yaml` holds an object
- **THEN** the assembled model has that object under the key `menu`

#### Scenario: The base file wins a conflict

- **WHEN** an object include and the base file both set the same key
- **THEN** the base file's value is the one in the assembled model

### Requirement: Per-item include directories each contribute one item

Each file under `quiqr/model/includes/singles/`, `includes/collections/` and
`includes/menus/` SHALL be read as one single, one collection, or one menu
section, and appended to the matching list.

#### Scenario: One single per file

- **WHEN** `includes/singles/about.yaml` describes a single
- **THEN** the assembled model has that single in its singles list

#### Scenario: Both include styles at once

- **WHEN** a site has both `includes/collections.yaml` and files under
  `includes/collections/`
- **THEN** every collection from both appears exactly once

### Requirement: Partials are merged into the item that names them

A single or collection carrying a `_mergePartial` value SHALL be merged with
the contents of the partial it names, with the item's own keys winning.

#### Scenario: A partial supplies the fields

- **WHEN** a collection declares only a key, a folder and a `_mergePartial`
- **AND** the partial declares the fields
- **THEN** the assembled collection has the fields from the partial

#### Scenario: The item overrides the partial

- **WHEN** both the item and its partial set a title
- **THEN** the item's title is the one in the assembled model

#### Scenario: A remote partial that was never cached

- **WHEN** a `_mergePartial` names an `http://` or `https://` location
- **AND** no cached copy is present
- **THEN** the error says the partial has not been fetched and names it
- **AND** the rest of the model is not silently assembled without it

### Requirement: Assembly is a pure read of a directory

Assembling a model SHALL read only from the site directory it is given, SHALL
make no network request, and SHALL write nothing.

#### Scenario: A read-only site directory

- **WHEN** a model is assembled from a directory the process cannot write to
- **THEN** assembly succeeds

#### Scenario: The same directory read twice

- **WHEN** the same unchanged site is assembled twice
- **THEN** both results are equal

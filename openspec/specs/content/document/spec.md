# content/document Specification

## Purpose
Defines what a content file is to ChatCMS: how it is split into front matter
and a body, and how the values in it correspond to the fields a model
declares.

## Requirements

### Requirement: A content file is front matter and a body

The system SHALL read a content file as a set of named values followed by a
Markdown body, in whichever of the supported formats the file uses, chosen by
the delimiter the file opens with.

#### Scenario: A file with YAML front matter

- **WHEN** a file opens with a line of three hyphens and closes the section
  with another
- **THEN** the text between them is read as YAML
- **AND** everything after it is the body

#### Scenario: A file with TOML front matter

- **WHEN** a file opens with a line of three plus signs
- **THEN** the text between the delimiters is read as TOML

#### Scenario: A file with no front matter

- **WHEN** a file opens with no delimiter
- **THEN** it has no named values
- **AND** the whole file is the body

#### Scenario: A file with front matter and no body

- **WHEN** a file closes its front matter and has nothing after it
- **THEN** the body is empty
- **AND** reading it does not fail

#### Scenario: Front matter that does not parse

- **WHEN** the text between the delimiters is not valid in its format
- **THEN** reading fails with an error naming the file and the position
- **AND** no partially read document is returned

### Requirement: Values correspond to the fields a model declares

Reading a document against a model SHALL place each value at the field that
owns it, and writing SHALL put it back in the same place.

#### Scenario: A plain field

- **WHEN** a model declares a field with a key
- **THEN** that field's value is the value of the matching named value

#### Scenario: A grouping field that nests

- **WHEN** a model declares a `section` without saying otherwise
- **THEN** its children's values live under the section's own key

#### Scenario: A grouping field that does not nest

- **WHEN** a `section` declares that it does not group its data
- **THEN** its children's values live beside it rather than under it

#### Scenario: A field that pulls its children under another name

- **WHEN** a model declares a `pull` naming a group
- **THEN** its children's values live under that name
- **AND** when it names none, they live under the pull's own key

#### Scenario: A repeating group

- **WHEN** a model declares an `accordion`
- **THEN** its value is a list, and each entry holds the children's values

#### Scenario: A repeating single field

- **WHEN** a model declares a `leaf-array`
- **THEN** its value is a list of the child field's values, not of objects

#### Scenario: Nesting to any depth

- **WHEN** grouping fields contain grouping fields
- **THEN** values are placed correctly at every level

### Requirement: The body is a value like any other

The Markdown body SHALL be readable and writable as the value of the field the
model uses for it, so that a form can edit it without knowing it is not part
of the front matter.

#### Scenario: Reading the body through a field

- **WHEN** a model declares a field for the body
- **THEN** that field's value is the Markdown after the front matter

#### Scenario: Writing the body through a field

- **WHEN** that field's value is changed and the document is written
- **THEN** the Markdown after the front matter is the new value
- **AND** the named values are unchanged

### Requirement: Nothing in a document is lost

A value present in a file that the model does not declare a field for SHALL be
preserved, and SHALL be written back in its original place.

#### Scenario: A value the model does not mention

- **WHEN** a file holds a named value no field claims
- **AND** the document is read and written with no change
- **THEN** that value is still there

#### Scenario: An edit beside an unclaimed value

- **WHEN** one field is changed and written
- **THEN** values no field claims are unchanged

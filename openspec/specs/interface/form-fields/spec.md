# interface/form-fields Specification

## Purpose
Defines what a person sees for each kind of field, and what the editor does
while they type.

## Requirements

### Requirement: Each kind of field has a control a person recognises

Every kind of field the model can declare SHALL be drawn with a control suited
to it.

#### Scenario: the kinds drawn

- **WHEN** a page is drawn
- **THEN** a short text, a long text, a number, a slider, a date, a choice, a
  yes or no, a colour, a list of values and a list of items SHALL each have a
  control of its own

#### Scenario: a long text says how long it is

- **WHEN** a field is a long text
- **THEN** how many characters it holds SHALL be shown

#### Scenario: a choice can be left unset

- **WHEN** a field is a choice
- **THEN** it SHALL be possible to choose nothing

### Requirement: A list can be added to and taken from

A list SHALL be able to grow and shrink, and each entry SHALL be nameable in
words a person recognises.

#### Scenario: adding

- **WHEN** somebody adds to a list
- **THEN** an empty entry SHALL appear, named after what the list holds

#### Scenario: removing

- **WHEN** somebody removes an entry
- **THEN** it SHALL go, and the rest SHALL stay in order

#### Scenario: an item of a list of items

- **WHEN** a list holds items with fields of their own
- **THEN** each SHALL be named by its first field, or by its position where
  that is empty
- **AND** SHALL open to show the fields inside it

### Requirement: A page's text is written in Markdown either way

The text editor SHALL write Markdown whether somebody uses the toolbar or
writes the marks themselves.

#### Scenario: the toolbar

- **WHEN** somebody uses the toolbar
- **THEN** the marks SHALL be put into the text, around what is selected or
  around a word to replace

#### Scenario: the source

- **WHEN** somebody asks for the source
- **THEN** exactly what will be saved SHALL be shown
- **AND** the toolbar SHALL be put away

### Requirement: Saving happens by itself

There SHALL be no Save button. An edit SHALL be written a short while after
typing stops, and what is said about it SHALL be one of three things: that
everything is saved, that it is saving, or that it has not been saved yet and
will be tried again.

#### Scenario: typing

- **WHEN** somebody types several letters quickly
- **THEN** one save SHALL be made rather than one for each letter

#### Scenario: saved

- **WHEN** a save works
- **THEN** it SHALL say that everything is saved

#### Scenario: not saved

- **WHEN** a save does not work
- **THEN** it SHALL say so, and that it will keep trying
- **AND** what was typed SHALL still be on the screen

#### Scenario: leaving the page

- **WHEN** somebody leaves the page with something not yet written
- **THEN** it SHALL be written

### Requirement: Nothing a content creator reads carries a technical word

Everything the editor says SHALL be in the words `docs/design/DESIGN.md`
section 7 allows.

#### Scenario: the editor's own words

- **WHEN** a page is open
- **THEN** nothing shown SHALL be a version control word

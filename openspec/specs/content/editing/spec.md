# content/editing Specification

## Purpose
Defines what a page looks like when it is opened for editing, and what happens
to the file when somebody changes something.

## Requirements

### Requirement: A page is addressed by what the model calls it

A page SHALL be addressed by the key the model gives it: a single page's own
key, or a collection's key and the item's name.

#### Scenario: a single page

- **WHEN** a page the model names as a single is asked for
- **THEN** it SHALL be read from the file the model names

#### Scenario: an item of a collection

- **WHEN** a collection's key and an item's name are asked for
- **THEN** the item SHALL be read from that collection's folder

#### Scenario: a name that reaches outside its collection

- **WHEN** an item's name climbs out of its folder, names a path, or is empty
- **THEN** the answer SHALL be that there is nothing there
- **AND** no file outside the collection SHALL be read

#### Scenario: a page the model does not name

- **WHEN** a key the model does not name is asked for
- **THEN** the answer SHALL be that there is nothing there

### Requirement: A page comes back as the controls to draw, with its values

Reading a page SHALL give what to draw and what is in it: each field the model
declares, the words to label it with, and the value the file holds.

#### Scenario: a field with a value

- **WHEN** a page is read
- **THEN** each field SHALL carry the value the file holds for it

#### Scenario: a field the file does not have

- **WHEN** the model declares a field the file has no value for
- **THEN** the field SHALL still be drawn, with no value

#### Scenario: a group of fields

- **WHEN** the model groups fields
- **THEN** the group SHALL carry the fields inside it
- **AND** SHALL carry no value of its own

#### Scenario: a list of items

- **WHEN** the model declares a list of items
- **THEN** it SHALL carry the items it holds
- **AND** SHALL carry the shape of one item, so that another can be added

#### Scenario: a field nobody is meant to see

- **WHEN** the model marks a field hidden
- **THEN** it SHALL NOT be drawn

### Requirement: A kind of field this version cannot draw is still shown

A field whose kind this version has no control for SHALL be shown as plain
text rather than left out, and SHALL say that it is.

#### Scenario: an unfamiliar kind

- **WHEN** the model declares a kind of field this version has no control for
- **THEN** it SHALL still be shown, with its value
- **AND** SHALL say that it is shown as plain text

#### Scenario: what is typed into it

- **WHEN** somebody types into such a field
- **THEN** what they typed SHALL be written exactly as typed

### Requirement: A page's own text is editable only where the model says so

A page's own text SHALL be editable where the model declares a field for it,
and readable but unwritable where it does not.

#### Scenario: a model that gives the text a field

- **WHEN** the model declares the field the reference implementation uses for
  a page's text
- **THEN** the page SHALL say its text can be edited
- **AND** the text SHALL NOT also be drawn as an ordinary field

#### Scenario: a model that does not

- **WHEN** the model declares no such field
- **THEN** the page SHALL say its text cannot be edited
- **AND** the text SHALL still be readable

#### Scenario: writing text that has no field

- **WHEN** text is written to a page whose model gives it no field
- **THEN** it SHALL be refused with a reason
- **AND** the file SHALL be unchanged

### Requirement: An edit changes only what it names

Writing an edit SHALL change the values it names and nothing else. The rest of
the file SHALL be byte for byte as it was.

#### Scenario: one field changed

- **WHEN** one field is written
- **THEN** the file SHALL differ only in that field's value

#### Scenario: the text changed

- **WHEN** only the page's text is written
- **THEN** the fields SHALL be exactly as they were

#### Scenario: a field cleared

- **WHEN** a field is written with nothing
- **THEN** the absent value the reference implementation writes SHALL be
  written

#### Scenario: a field the page does not have

- **WHEN** a value is written to a field the page's model does not declare
- **THEN** it SHALL be refused
- **AND** nothing SHALL be written

### Requirement: Every save makes a version and asks for a sync

Writing an edit SHALL make a version of the site and ask for it to be brought
level with its remote.

#### Scenario: a save

- **WHEN** an edit is written
- **THEN** a version SHALL be made, attributed to the person who made the edit
- **AND** the site SHALL be queued to be brought level with its remote

#### Scenario: the queue is not working

- **WHEN** the site cannot be queued for a sync
- **THEN** the save SHALL still have happened
- **AND** the person SHALL NOT be told that their edit failed

### Requirement: A model that cannot be read says where the mistake is

A site whose model has a mistake in it is different from a site with no model,
and SHALL be reported differently.

#### Scenario: no model at all

- **WHEN** a site carries no model
- **THEN** the answer SHALL say there are no forms yet

#### Scenario: a model with a mistake

- **WHEN** a site's model cannot be read
- **THEN** the answer SHALL carry what is wrong with it, naming the file

### Requirement: Only somebody allowed to edit may write

Whether somebody may read a page and whether they may change it SHALL both be
decided by the policy function.

#### Scenario: a viewer reads and tries to write

- **WHEN** somebody who may read but not change the site opens a page
- **THEN** they SHALL be shown it
- **AND** an attempt to write SHALL be refused, leaving the file unchanged

#### Scenario: somebody with no part in the project

- **WHEN** somebody with no part in the project asks for a page
- **THEN** they SHALL be told there is nothing there

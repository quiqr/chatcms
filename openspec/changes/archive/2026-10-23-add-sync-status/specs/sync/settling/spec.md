## Purpose

Defines what happens when two people changed the same thing, so that choosing
is about a page rather than about a file.

## ADDED Requirements

### Requirement: A difference is asked about field by field

Where two people changed the same page, each field they both changed SHALL be
a question of its own, named the way the model names it.

#### Scenario: one field changed by both

- **WHEN** two people changed the same field
- **THEN** it SHALL be a question carrying both answers
- **AND** the question SHALL name the page and the field in the site's words

#### Scenario: a field only one person changed

- **WHEN** only one person changed a field
- **THEN** it SHALL NOT be asked about

#### Scenario: the page's own text

- **WHEN** both changed the page's text
- **THEN** it SHALL be a question like any other field

#### Scenario: comparing

- **WHEN** the two answers are compared
- **THEN** the words that differ SHALL be marked

### Requirement: A file that cannot be shown field by field is one choice

A file that does not hold words the model names SHALL be a choice between one
whole version and the other.

#### Scenario: a picture

- **WHEN** two people changed the same picture
- **THEN** it SHALL be a choice between whole versions
- **AND** SHALL NOT be shown as text

#### Scenario: a page whose difference cannot be named

- **WHEN** two versions of a page differ somewhere no field covers
- **THEN** it SHALL be a choice between whole versions

### Requirement: Doing nothing keeps the work of whoever is here

A field nobody answers about SHALL keep what was written here, so that doing
nothing never throws away the work of the person doing it.

#### Scenario: a field left unanswered

- **WHEN** an answer says nothing about a field
- **THEN** what was written here SHALL be kept

#### Scenario: taking the other version

- **WHEN** an answer asks for the other version of a field
- **THEN** that value SHALL be written into the page
- **AND** every other field SHALL keep what was written here

### Requirement: Everything has to be answered before anything is written

An exchange SHALL be settled whole or not at all.

#### Scenario: a page left out

- **WHEN** an answer leaves out a page that two people changed
- **THEN** it SHALL be refused with a reason
- **AND** nothing SHALL be written

### Requirement: Giving up puts the site back

Giving up on an unfinished question SHALL restore the site as it was.

#### Scenario: starting again

- **WHEN** somebody gives up on the question
- **THEN** the site SHALL be put back as it was

### Requirement: Only somebody allowed to edit may settle anything

Whether somebody may settle what two people changed SHALL be decided by the
policy function.

#### Scenario: a viewer tries

- **WHEN** somebody who may read but not change the site tries to bring in
  changes, answer, or give up
- **THEN** each SHALL be refused
- **AND** nothing SHALL be written

# content/pages Specification

## Purpose
Defines making, renaming, copying and removing a page, and what the files are
called.

## Requirements

### Requirement: A file is named the way the old Quiqr names it

A page's file SHALL be named from its title: small letters, digits and hyphens,
with accents folded to the letter they are made from.

#### Scenario: an ordinary title

- **WHEN** a page is called "Spring prices"
- **THEN** its file SHALL be named `spring-prices` with the collection's
  extension

#### Scenario: a title with accents

- **WHEN** a title carries accented letters
- **THEN** they SHALL be folded to the letters they are made from

#### Scenario: a title that would become a path

- **WHEN** a title carries a slash or dots
- **THEN** the name SHALL carry none of them
- **AND** the file SHALL be inside the collection's own folder

#### Scenario: a title with nothing to name a file from

- **WHEN** a title has no letters or digits at all
- **THEN** it SHALL be refused with a reason

### Requirement: A page is made in a collection

A page SHALL be made only inside a collection the model declares, holding its
title and nothing else.

#### Scenario: making one

- **WHEN** a page is made in a collection
- **THEN** a file SHALL be written holding its title and nothing else
- **AND** a version SHALL be made and a sync asked for

#### Scenario: a name already taken

- **WHEN** a page would have the name of one that is there
- **THEN** it SHALL be refused, saying which name
- **AND** nothing SHALL be written

#### Scenario: a group that cannot grow

- **WHEN** a page is asked for in the pages group, or in a collection that is
  not there
- **THEN** the answer SHALL be that there is nothing there

### Requirement: Renaming moves the file and the title together

Renaming SHALL move the file and change the title inside it, because to a
content creator those are one thing.

#### Scenario: renaming

- **WHEN** a page is renamed
- **THEN** its file SHALL be written under the new name and the old one removed
- **AND** the title inside it SHALL be the new one

#### Scenario: renaming to the same name

- **WHEN** a page is renamed to what it is already called
- **THEN** nothing SHALL be written

#### Scenario: a name already taken

- **WHEN** the new name is one another page has
- **THEN** it SHALL be refused and nothing SHALL change

### Requirement: A copy leaves the first one alone

Copying SHALL make a second page and leave the first exactly as it was.

#### Scenario: copying

- **WHEN** a page is copied
- **THEN** a second file SHALL be written under the new name
- **AND** the page it was copied from SHALL still be there

### Requirement: Removing says what will happen first

Removing SHALL be confirmed, and the confirmation SHALL say what happens to
the page and how it can be brought back.

#### Scenario: what somebody is told

- **WHEN** removing is offered
- **THEN** it SHALL say that the page goes at the next publish and can be
  brought back

#### Scenario: removing

- **WHEN** a page is removed
- **THEN** the file SHALL go and a version SHALL be made
- **AND** a sync SHALL be asked for

### Requirement: A page the model names is not renamed or removed here

A page the model names SHALL NOT be renamed or removed through the browser,
because its name is part of how the site is built.

#### Scenario: trying

- **WHEN** renaming or removing is asked for a page the model names
- **THEN** it SHALL be refused, saying that its name is part of how the site
  is built
- **AND** nothing SHALL change

### Requirement: Only somebody allowed to edit may change what is there

Whether somebody may change what a site holds SHALL be decided by the policy
function.

#### Scenario: a viewer tries

- **WHEN** somebody who may read but not change the site tries to make,
  rename, copy or remove a page
- **THEN** each SHALL be refused
- **AND** the site SHALL be unchanged

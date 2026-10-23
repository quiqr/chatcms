# media/library Specification

## Purpose
Defines what a site's pictures are to the people editing it.

## Requirements

### Requirement: Every picture in the site is listed

The library SHALL list the pictures in the site's own picture folder, in the
same order every time, with what is known about each.

#### Scenario: what is listed

- **WHEN** the library is asked for
- **THEN** each picture SHALL carry its name, where it lives, how large it is
  and what it shows

#### Scenario: a site that already has a folder of its own

- **WHEN** a site already keeps pictures somewhere a Hugo site keeps them
- **THEN** that folder SHALL be used rather than a new one

#### Scenario: a site with no pictures

- **WHEN** a site has none
- **THEN** the library SHALL be empty rather than an error

### Requirement: What uses a picture is found by looking

Which pages refer to a picture SHALL be worked out by reading the content, not
from anything the app wrote down.

#### Scenario: a picture on a page

- **WHEN** a page refers to a picture, however its address is written
- **THEN** that page SHALL be listed as using it

#### Scenario: a picture nothing refers to

- **WHEN** nothing refers to a picture
- **THEN** it SHALL be listed as used nowhere

### Requirement: A description is kept with the site

What a picture shows SHALL be written into the site, so that it travels with
it.

#### Scenario: writing one

- **WHEN** somebody describes a picture
- **THEN** it SHALL be saved into the site and a version made

#### Scenario: the file is written the same way every time

- **WHEN** the same descriptions are written twice
- **THEN** the file SHALL be byte for byte the same

#### Scenario: an empty description

- **WHEN** a description is emptied
- **THEN** nothing SHALL be kept for that picture

#### Scenario: a picture that goes

- **WHEN** a picture is removed
- **THEN** its description SHALL go with it

### Requirement: Removing warns when something still uses it

Removing a picture pages still refer to SHALL be refused unless whoever asked
says to do it anyway.

#### Scenario: a picture in use

- **WHEN** removing a picture something refers to is asked for
- **THEN** it SHALL be refused, saying how many pages use it
- **AND** nothing SHALL be removed

#### Scenario: asked for again

- **WHEN** removing is asked for again, saying to do it anyway
- **THEN** the picture SHALL go

#### Scenario: a picture nothing uses

- **WHEN** a picture nothing refers to is removed
- **THEN** it SHALL go without asking

### Requirement: Looking needs only the right to read, changing needs more

Seeing the pictures SHALL need the right to read the project; adding,
describing and removing SHALL need the right to change its media.

#### Scenario: a viewer

- **WHEN** somebody who may read but not change the site opens the library
- **THEN** they SHALL see every picture and be able to look at one
- **AND** adding, describing and removing SHALL each be refused

## Purpose

Defines what ChatCMS will take as a picture, what it calls it, and what it
gives back.

## ADDED Requirements

### Requirement: What a picture is comes from its own bytes

What kind of picture a file is SHALL be decided by reading it, never by its
name, because a name comes from outside and a misnamed file would be published
as a broken picture.

#### Scenario: the kinds taken

- **WHEN** a PNG, JPEG, GIF, WebP or SVG arrives
- **THEN** it SHALL be taken

#### Scenario: something that is not a picture

- **WHEN** a file that is not a picture arrives, whatever it is called
- **THEN** it SHALL be refused, saying which kinds are taken
- **AND** nothing SHALL be written

#### Scenario: an empty file

- **WHEN** a file with nothing in it arrives
- **THEN** it SHALL be refused

### Requirement: A picture is named the way everything else is

A picture's name SHALL be made the way a page's is: small letters, digits and
hyphens, with the extension taken from what the bytes turned out to be.

#### Scenario: a name with spaces and capitals

- **WHEN** a file called "Our New Shop.PNG" arrives
- **THEN** it SHALL be named `our-new-shop.png`

#### Scenario: a name that says the wrong kind

- **WHEN** a file's name says one kind and its bytes say another
- **THEN** the name SHALL carry what the bytes say

#### Scenario: a name with nothing in it

- **WHEN** a file's name has no letters or digits
- **THEN** it SHALL be refused

### Requirement: A picture already there is not replaced

A picture arriving with the name of one already in the site SHALL be put
beside it rather than over it, because replacing would change every page that
uses the first.

#### Scenario: a second one of the same name

- **WHEN** a picture arrives with the name of one already there
- **THEN** it SHALL be put beside it under a name of its own
- **AND** the one already there SHALL be untouched

### Requirement: A picture that is too large is refused in words

A picture larger than a site will carry SHALL be refused with its size said
in words.

#### Scenario: too large

- **WHEN** a picture larger than a site will take arrives
- **THEN** it SHALL be refused
- **AND** the refusal SHALL say its size and the largest allowed, in words
  rather than in bytes

### Requirement: A small copy is made for showing many at once

A small copy SHALL be made the first time one is asked for and kept, named
after the bytes it was made from.

#### Scenario: the same picture twice

- **WHEN** a small copy of the same picture is asked for twice
- **THEN** the same bytes SHALL come back
- **AND** the work SHALL be done once

#### Scenario: throwing the copies away

- **WHEN** the kept copies are deleted
- **THEN** nothing SHALL be lost that cannot be made again from the site alone

#### Scenario: a picture that cannot be resized

- **WHEN** a kind that cannot be resized is asked for small
- **THEN** it SHALL be given back whole rather than reported as broken

### Requirement: A picture is served without inviting a browser to guess

Serving a picture SHALL state its kind and SHALL tell the browser not to
guess a different one.

#### Scenario: serving one

- **WHEN** a picture is served
- **THEN** its kind SHALL be stated
- **AND** the browser SHALL be told not to guess a different one

### Requirement: A name from outside cannot reach another part of the site

A picture's name coming from outside SHALL be checked before it is used to
reach a file.

#### Scenario: a name that climbs

- **WHEN** a name carries a path, dots that climb, or nothing at all
- **THEN** reading, describing and removing SHALL each answer that there is
  nothing there

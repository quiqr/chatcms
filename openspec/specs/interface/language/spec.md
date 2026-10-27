# interface/language Specification

## Purpose
TBD - created by archiving change add-dutch-and-one-place-for-strings. Update Purpose after archive.

## Requirements

### Requirement: Every string is in one place

No text a person can read SHALL be written where it is shown. Every string
SHALL come from the catalogue, named by a key.

#### Scenario: A string written in a screen

- **WHEN** a screen is given text rather than a key
- **THEN** the frontend check SHALL fail, naming the file and the text

#### Scenario: Text that is not a string

- **WHEN** what a screen holds is an address, a class name, an identifier or
  a value the person themselves typed
- **THEN** the check SHALL leave it alone

### Requirement: Another language is a file typed against the first

A language SHALL be a file holding exactly the keys the English holds. A
missing or unknown key SHALL fail the build rather than appear on a screen.

#### Scenario: A key nobody translated

- **WHEN** a language is missing a key the English has
- **THEN** the type check SHALL fail, naming the key

#### Scenario: A key that is not in the catalogue

- **WHEN** a screen asks for a key that does not exist
- **THEN** the type check SHALL fail

### Requirement: The interface is readable in Dutch

Every screen a person can reach SHALL be readable in Dutch, with no English
left in it.

#### Scenario: Switching

- **WHEN** somebody chooses Dutch
- **THEN** every screen SHALL be in Dutch at once, without signing out

#### Scenario: What the server refused

- **WHEN** the server refuses something and says why
- **THEN** the screen SHALL say it in the chosen language when it knows the
  refusal, and in the server's own words when it does not

#### Scenario: The suite in Dutch

- **WHEN** the acceptance suite runs in Dutch
- **THEN** the screens a content creator uses SHALL be in Dutch, which is how
  a string nobody moved fails for us rather than for a customer

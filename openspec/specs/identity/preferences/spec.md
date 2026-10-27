# identity/preferences Specification

## Purpose
What a person may choose about how the application behaves, and where that
choice is kept.

## Requirements

### Requirement: A preference is kept with the account

What somebody chooses about how the application behaves SHALL be kept with
their account, so that it is in force wherever they sign in.

#### Scenario: Chosen on one machine

- **WHEN** somebody chooses an appearance and then signs in elsewhere
- **THEN** the same appearance SHALL be in force there

#### Scenario: Never chosen

- **WHEN** somebody has chosen nothing
- **THEN** the application SHALL follow what their machine asks for
- **AND** no screen SHALL require a choice to be made first

#### Scenario: Somebody else's preference

- **WHEN** somebody asks to change another person's preference
- **THEN** it SHALL be refused

### Requirement: There is a dark appearance

The interface SHALL have a dark appearance as well as a light one, and SHALL
follow the machine when neither is chosen.

#### Scenario: Choosing dark

- **WHEN** somebody chooses the dark appearance
- **THEN** every screen SHALL be drawn in it

#### Scenario: Following the machine

- **WHEN** somebody has chosen neither
- **THEN** the appearance SHALL be what the machine asks for
- **AND** SHALL change with it

#### Scenario: No colour written by hand

- **WHEN** the interface is checked
- **THEN** no colour outside the named tokens SHALL appear in either
  appearance

### Requirement: A date is written one way

Every screen SHALL write a date and a time through one function, in the
style the design asks for and in the language that is in force.

#### Scenario: Something that happened today

- **WHEN** a time today is shown
- **THEN** it SHALL be written as the design asks rather than as a full date

#### Scenario: Every screen

- **WHEN** any screen shows a date
- **THEN** it SHALL use that function, and no screen SHALL format one itself

#### Scenario: In Dutch

- **WHEN** the language is Dutch
- **THEN** the day, the month and the words around them SHALL be Dutch

### Requirement: A language is kept with the account

Somebody SHALL be able to choose the language of the interface, and that
choice SHALL be kept with their account the way the appearance is. Choosing
nothing SHALL mean the language the browser asks for.

#### Scenario: Choosing one

- **WHEN** somebody chooses Dutch
- **THEN** every screen SHALL be in Dutch at once
- **AND** it SHALL still be Dutch when they sign in on another machine

#### Scenario: Never chosen

- **WHEN** somebody has chosen no language
- **THEN** the interface SHALL be in the language their browser asks for,
  where that is one the interface has

#### Scenario: A language the interface does not have

- **WHEN** a browser asks for a language the interface does not have
- **THEN** English SHALL be used

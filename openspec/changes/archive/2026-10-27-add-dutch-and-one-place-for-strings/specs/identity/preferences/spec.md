## ADDED Requirements

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

## MODIFIED Requirements

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

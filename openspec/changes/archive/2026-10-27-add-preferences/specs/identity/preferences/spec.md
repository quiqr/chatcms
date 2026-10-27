## Purpose

What a person may choose about how the application behaves, and where that
choice is kept.

## ADDED Requirements

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
style the design asks for.

#### Scenario: Something that happened today

- **WHEN** a time today is shown
- **THEN** it SHALL be written as the design asks rather than as a full date

#### Scenario: Every screen

- **WHEN** any screen shows a date
- **THEN** it SHALL use that function, and no screen SHALL format one itself

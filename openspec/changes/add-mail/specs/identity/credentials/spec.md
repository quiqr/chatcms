## Purpose

An invitation and a reset are now sent, not only issued.

## ADDED Requirements

### Requirement: A link that is issued is sent

Issuing an invitation or a password reset SHALL hand the person a message
carrying the link, through whatever sending is configured.

#### Scenario: An invitation

- **WHEN** somebody is invited
- **THEN** a message SHALL be addressed to them with the link in it

#### Scenario: A reset that was asked for

- **WHEN** somebody asks to set a new password and has an account
- **THEN** a message SHALL be addressed to them with the link in it

#### Scenario: A reset for an address nobody holds

- **WHEN** somebody asks for an address with no account
- **THEN** no message SHALL be sent
- **AND** the screen SHALL say exactly what it says when there is one

### Requirement: The link has somewhere to land

The address a sent link points at SHALL be a screen that sets a password.

#### Scenario: Following an invitation

- **WHEN** somebody opens the link from their invitation
- **THEN** they SHALL be asked to choose a password
- **AND** setting one SHALL sign them in

#### Scenario: A link with nothing in it

- **WHEN** the address carries no link
- **THEN** the screen SHALL say so and say what to do about it

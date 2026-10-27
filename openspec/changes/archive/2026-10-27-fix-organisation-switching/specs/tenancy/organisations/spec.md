## ADDED Requirements

### Requirement: Somebody in more than one organisation can switch between them

Somebody who belongs to more than one organisation SHALL be able to go from
one to another from any screen that shows the switcher, without typing an
address.

#### Scenario: Choosing the other one

- **WHEN** somebody chooses another organisation in the switcher
- **THEN** they SHALL be taken to that organisation's projects
- **AND** that organisation's sites SHALL be what is listed

#### Scenario: Where you are

- **WHEN** the switcher is open
- **THEN** the organisation whose screen it is SHALL be marked as the one
  being looked at

#### Scenario: Somebody in one organisation

- **WHEN** somebody belongs to one organisation
- **THEN** the switcher SHALL still say which one, and offer nothing else

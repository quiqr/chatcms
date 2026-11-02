## ADDED Requirements

### Requirement: An organisation has screens of its own

An organisation SHALL be reachable as a thing in itself, not only through one
of the projects inside it. Whoever administers it SHALL be able to see what it
holds, who is in it, and how full it is, from screens that belong to the
organisation.

#### Scenario: Reaching the area

- **WHEN** somebody who administers an organisation opens the switcher
- **THEN** they SHALL be offered a way to manage that organisation
- **AND** somebody who merely belongs to it SHALL NOT be

#### Scenario: Somebody who merely belongs

- **WHEN** somebody who is not an admin of the organisation asks for its
  screens
- **THEN** it SHALL be refused by the permission function

#### Scenario: What the area says about itself

- **WHEN** the area is open
- **THEN** it SHALL say which organisation it is, how many members and sites
  it holds, how full it is against any limits, and who sets those limits

### Requirement: An organisation's members are managed from the organisation

Who belongs to an organisation SHALL be managed from the organisation's own
screens. Somebody taken out of an organisation SHALL lose the roles they held
on its projects.

#### Scenario: Inviting somebody

- **WHEN** an organisation admin invites an address with a role in the
  organisation
- **THEN** that person SHALL be sent the way in, and SHALL hold that role
- **AND** the seat they take SHALL count against the limit at once

#### Scenario: No seats left

- **WHEN** every seat is in use
- **THEN** the screen SHALL say so, and SHALL NOT offer the form
- **AND** the server SHALL refuse an invitation anyway

#### Scenario: Taking somebody out

- **WHEN** somebody is taken out of an organisation
- **THEN** they SHALL lose their roles on its projects as well
- **AND** nothing they wrote SHALL be changed or removed

### Requirement: An organisation's sites are listed as a list

The organisation's sites SHALL be shown as a list that reflows rather than a
table, with what each one is, how many people reach it, when it was last
published, and what state it is in.

#### Scenario: Opening one

- **WHEN** an organisation admin opens a site from the list
- **THEN** they SHALL reach it as its implementer, without being given a role
  on it

#### Scenario: What cannot be done yet

- **WHEN** the list offers an act that is not built
- **THEN** it SHALL be offered disabled, saying so, rather than left out or
  offered and broken

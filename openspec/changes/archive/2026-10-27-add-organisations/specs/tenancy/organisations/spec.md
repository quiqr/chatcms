## Purpose

Gives an organisation a name that can change and people who can be seen.

## ADDED Requirements

### Requirement: An organisation can be renamed

An organisation's name SHALL be changeable by an administrator of the
installation, and the new name SHALL be what every screen shows from then on.
Its address SHALL NOT change with it, because links already written would
stop working.

#### Scenario: Renaming it

- **WHEN** an administrator gives an organisation a new name
- **THEN** it SHALL be called that
- **AND** its address SHALL be what it was

#### Scenario: Renaming it as its owner

- **WHEN** the organisation's own owner asks to rename it
- **THEN** it SHALL be refused

#### Scenario: A name with nothing in it

- **WHEN** the new name is empty or only spaces
- **THEN** it SHALL be refused, and the old name SHALL stand

### Requirement: The people of an organisation can be seen and changed

Somebody who owns an organisation SHALL see who is in it and what each holds,
and SHALL be able to change a role or take somebody out.

#### Scenario: Its owner

- **WHEN** an organisation's owner asks who is in it
- **THEN** every person SHALL be listed with the role they hold there

#### Scenario: Somebody who is merely in it

- **WHEN** a member who does not own it asks to change a role
- **THEN** it SHALL be refused

#### Scenario: Somebody with no part in it

- **WHEN** somebody who does not belong asks who is in it
- **THEN** the answer SHALL be that there is nothing there

#### Scenario: The last owner

- **WHEN** a change would leave an organisation with no owner
- **THEN** it SHALL be refused, and the message SHALL say why

### Requirement: A screen knows which organisation it is showing

Every screen SHALL take the organisation it is showing from the address, and
SHALL NOT assume one.

#### Scenario: An address naming an organisation somebody is not in

- **WHEN** somebody opens an address for an organisation they do not belong
  to and do not administer
- **THEN** the answer SHALL be that there is nothing there
- **AND** they SHALL NOT be sent to one they do belong to

#### Scenario: Two organisations

- **WHEN** somebody belongs to two and switches between them
- **THEN** each SHALL show its own projects
- **AND** nothing from one SHALL appear in the other

#### Scenario: Belonging to none

- **WHEN** somebody who belongs to no organisation opens their projects
- **THEN** they SHALL be told so, and told what to do about it
- **AND** an administrator SHALL additionally be offered to start one

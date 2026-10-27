## Purpose

An administrator can read an organisation in full and remove one.

## ADDED Requirements

### Requirement: An administrator reads an organisation whole

An administrator SHALL be able to read one organisation with everything
about it: its name, its address, who belongs to it and what they hold
there, and the projects it holds.

#### Scenario: Reading one

- **WHEN** an administrator opens an organisation
- **THEN** its people and its projects SHALL be shown
- **AND** an organisation they do not belong to SHALL be readable just the
  same

#### Scenario: Somebody who does not administer

- **WHEN** somebody who does not administer the installation asks
- **THEN** it SHALL be refused

### Requirement: An administrator removes an organisation

Removing an organisation SHALL need administration of the installation, and
SHALL NOT change anything in any git repository.

#### Scenario: Removing an empty one

- **WHEN** an administrator removes an organisation that holds no projects
- **THEN** it SHALL be gone
- **AND** the act SHALL be in the administration record

#### Scenario: One that still holds projects

- **WHEN** an organisation still holds projects
- **THEN** removing it SHALL be refused
- **AND** the message SHALL name what is in the way

#### Scenario: What is in git

- **WHEN** an organisation is removed
- **THEN** every git remote its projects used SHALL hold what it held

## Purpose

A project's name, where its site lives and how it is published can be changed
after it has been set up, and a project can be removed.

## ADDED Requirements

### Requirement: What setup decided can be changed afterwards

A project's name, the address shown to customers, where its site lives and
how it is published SHALL each be changeable after the project is set up, by
somebody allowed to change its settings.

#### Scenario: Renaming it

- **WHEN** somebody allowed to change settings gives a project a new name
- **THEN** it SHALL be called that everywhere it is shown
- **AND** the address it is reached at SHALL NOT change

#### Scenario: An editor

- **WHEN** somebody who may only edit content opens the settings
- **THEN** it SHALL be refused by the permission rather than by a hidden link

#### Scenario: Changing where the site lives

- **WHEN** the address of the site in git is changed
- **THEN** the connection SHALL be tested before it is saved
- **AND** the work already brought in SHALL stay

#### Scenario: Changing how it is published

- **WHEN** the line it publishes to, or whether it sends source or built
  pages, is changed
- **THEN** what the next publish will do differently SHALL be said before it
  is saved
- **AND** the next publish SHALL do it

### Requirement: A project can be removed without touching its site

Removing a project SHALL remove it from ChatCMS and SHALL NOT change anything
in the git repository it was reading.

#### Scenario: Removing one

- **WHEN** somebody allowed to remove a project confirms it
- **THEN** the project SHALL be gone from ChatCMS
- **AND** the remote SHALL hold what it held

#### Scenario: What the confirmation says

- **WHEN** the confirmation is shown
- **THEN** it SHALL name what is removed and say that the site itself is not

#### Scenario: Somebody not allowed

- **WHEN** somebody who may not remove a project asks to
- **THEN** it SHALL be refused

### Requirement: One answer says how much is waiting

How many changes are waiting to be published SHALL come from the same answer
wherever it is shown.

#### Scenario: The count beside the screen

- **WHEN** the publish screen says how many changes are waiting
- **THEN** the count shown on the button that leads to it SHALL be the same

#### Scenario: Nothing waiting

- **WHEN** nothing is waiting
- **THEN** no count SHALL be shown

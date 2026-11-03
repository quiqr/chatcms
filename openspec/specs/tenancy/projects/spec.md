# tenancy/projects Specification

## Purpose
A project's name, where its site lives and how it is published can be changed
after it has been set up, and a project can be removed.

## Requirements

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

### Requirement: A site can be duplicated within its organisation

A site SHALL be able to be copied into the same organisation. The copy SHALL
have a name of its own, and SHALL share nothing with the original.

#### Scenario: The copy is its own site

- **WHEN** a site is duplicated
- **THEN** the copy SHALL hold the pages, the media and the content model the
  original had
- **AND** editing either SHALL leave the other as it was

#### Scenario: Where the copy publishes

- **WHEN** a site is duplicated without asking for its publishing settings
- **THEN** the copy SHALL publish nowhere until somebody says where
- **AND** publishing the copy SHALL NOT change the original's website

#### Scenario: The same name

- **WHEN** the copy is given the name the original has
- **THEN** it SHALL be refused, saying the name is taken

#### Scenario: No room for it

- **WHEN** the organisation is at its limit of sites
- **THEN** duplicating SHALL be refused before anything is made

### Requirement: A site can be moved to another organisation

A site SHALL be able to move to another organisation. Whoever moves it SHALL
administer both, unless they administer the installation, in which case any
organisation is allowed.

#### Scenario: What moves with it

- **WHEN** a site is moved
- **THEN** its files, its history and where it publishes SHALL be unchanged
- **AND** the people who held a role on it SHALL keep it
- **AND** those people SHALL be members of the organisation it moved to

#### Scenario: An organisation somebody does not administer

- **WHEN** somebody asks to move a site into an organisation they only belong
  to
- **THEN** it SHALL be refused

#### Scenario: No room for the people

- **WHEN** the people who would come with the site would take the
  destination over its limit of members
- **THEN** the move SHALL be refused before anything changes

### Requirement: A site can be downloaded and archived

A site SHALL be able to be downloaded as a zip of its files, and to be
archived.

#### Scenario: The zip

- **WHEN** a site is downloaded
- **THEN** the zip SHALL hold the site's files as they are, and nothing of
  how git keeps them

#### Scenario: Archiving

- **WHEN** a site is archived
- **THEN** it SHALL stop counting against the organisation's limit of sites
- **AND** nothing SHALL be removed, and what is published SHALL stay published

#### Scenario: Coming back

- **WHEN** an archived site is brought back
- **THEN** it SHALL count again, and be where it was

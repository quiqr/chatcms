## ADDED Requirements

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

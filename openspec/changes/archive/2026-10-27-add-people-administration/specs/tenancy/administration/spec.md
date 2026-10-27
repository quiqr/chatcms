## ADDED Requirements

### Requirement: An administrator sees every person

An administrator SHALL be able to see every person on the installation: their
name, their address, whether their account is on, when they were last here,
and where they belong. Where somebody belongs SHALL be read with the list
rather than with a request for each person.

#### Scenario: Listing them

- **WHEN** an administrator asks who is on the installation
- **THEN** every person SHALL be listed, including people in no organisation
- **AND** each SHALL carry the organisations they are in and the sites they
  hold a role on

#### Scenario: Listing them as anybody else

- **WHEN** somebody who does not administer the installation asks
- **THEN** it SHALL be refused, whatever they hold on any organisation

#### Scenario: Finding somebody

- **WHEN** an administrator searches by part of a name or an address
- **THEN** only the people that match SHALL be listed

### Requirement: An administrator opens one person

An administrator SHALL be able to open one person and see everything that
person can reach, and nothing about anybody else.

#### Scenario: Opening one

- **WHEN** an administrator opens a person
- **THEN** their name, address, whether their account is on, when they were
  last here, their organisations with the role held in each, and their sites
  with the role held on each SHALL be answered

#### Scenario: Turning an account off from there

- **WHEN** an administrator turns that person's account off
- **THEN** the person SHALL no longer be able to sign in
- **AND** the sessions they already had SHALL stop

#### Scenario: Somebody who is not there

- **WHEN** an administrator opens an address no person has
- **THEN** it SHALL say so rather than answer an empty person

### Requirement: An administrator makes a person

An administrator SHALL be able to make a person from an address and a name,
without choosing an organisation or a site first. The person SHALL be sent
the same way in that every other invitation sends, and no password SHALL be
set by anybody but the person themselves.

#### Scenario: Making one

- **WHEN** an administrator makes a person
- **THEN** that person SHALL exist with no password
- **AND** SHALL be sent a link with which to choose one

#### Scenario: An address that is already somebody

- **WHEN** an administrator makes a person at an address that already has an
  account
- **THEN** it SHALL be refused, naming the address, and nothing SHALL change
  about that account

#### Scenario: Making one as anybody else

- **WHEN** somebody who does not administer the installation asks to make a
  person
- **THEN** it SHALL be refused

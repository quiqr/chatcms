# tenancy/administration Specification

## Purpose
Defines who holds the installation itself, what that lets them do, what it
deliberately does not, and how the first one comes to exist.

## Requirements

### Requirement: An administrator holds the installation

An installation SHALL know which people administer it. Administering SHALL be
a property of a person rather than a role on any organisation, and SHALL be
decided by the same function that decides every other permission.

#### Scenario: Somebody who is not an administrator

- **WHEN** somebody who does not administer the installation asks to do
  something only an administrator may
- **THEN** it SHALL be refused
- **AND** the refusal SHALL come from the permission function rather than from
  a check written beside it

#### Scenario: Nobody at all

- **WHEN** somebody not signed in asks
- **THEN** it SHALL be refused

### Requirement: Only an administrator shapes organisations

Creating an organisation, renaming one, and removing one SHALL each need
administration of the installation.

#### Scenario: Creating one

- **WHEN** an administrator creates an organisation
- **THEN** it SHALL exist
- **AND** the administrator SHALL NOT thereby become one of its people

#### Scenario: Anybody else creating one

- **WHEN** somebody who is not an administrator asks to create an organisation
- **THEN** it SHALL be refused, whatever they hold on any other organisation

#### Scenario: Putting an owner in

- **WHEN** an administrator gives somebody ownership of an organisation
- **THEN** that person SHALL hold it
- **AND** SHALL see the organisation among their own

### Requirement: An administrator sees every organisation

An administrator SHALL be able to see every organisation, who is in each, and
which projects each holds.

#### Scenario: Listing them

- **WHEN** an administrator asks what organisations exist
- **THEN** every one SHALL be listed, including those they are not in

#### Scenario: Listing them as anybody else

- **WHEN** somebody who is not an administrator asks
- **THEN** only the organisations they belong to SHALL be listed

### Requirement: Administering is not editing

Administering the installation SHALL NOT by itself grant any permission over
a project's content, media, publishing or history.

#### Scenario: Reading a customer's content

- **WHEN** an administrator who belongs to no organisation asks to read a
  project's content
- **THEN** it SHALL be refused

#### Scenario: Publishing a customer's site

- **WHEN** such an administrator asks to publish
- **THEN** it SHALL be refused

#### Scenario: Joining first

- **WHEN** an administrator puts themselves into an organisation
- **THEN** they SHALL hold what that membership grants and no more
- **AND** the organisation's own people SHALL see them in its list of people

### Requirement: The first administrator is appointed by configuration

The server SHALL take an email address from its configuration, as a value or
from a file, and SHALL make that person an administrator when it starts.

#### Scenario: A person with that address exists

- **WHEN** the server starts and that person exists and is not yet an
  administrator
- **THEN** they SHALL become one
- **AND** the log SHALL say so

#### Scenario: Starting again

- **WHEN** the server starts again with the same address
- **THEN** nothing SHALL change
- **AND** the log SHALL NOT claim to have appointed anybody

#### Scenario: No such person yet

- **WHEN** no person has that address
- **THEN** the server SHALL start anyway
- **AND** SHALL say that it could not appoint anybody and why

#### Scenario: No address configured

- **WHEN** no address is configured
- **THEN** the server SHALL start and appoint nobody

### Requirement: Administrators appoint each other, and the last one stays

An administrator SHALL be able to make another person an administrator and to
take it away, except that the last administrator SHALL NOT be removed.

#### Scenario: Appointing somebody

- **WHEN** an administrator appoints another person
- **THEN** that person SHALL administer the installation

#### Scenario: Removing the last one

- **WHEN** removing would leave the installation with no administrator
- **THEN** it SHALL be refused, and the message SHALL say why

#### Scenario: Standing down

- **WHEN** an administrator removes themselves while another remains
- **THEN** it SHALL be allowed

### Requirement: What an administrator does is recorded

Every act that needs administration SHALL be recorded with who did it, what
they did, what it was done to, and when.

#### Scenario: An organisation is made

- **WHEN** an administrator creates an organisation
- **THEN** the record SHALL hold who, what and when

#### Scenario: Somebody is appointed

- **WHEN** an administrator appoints or removes an administrator
- **THEN** the record SHALL hold it

#### Scenario: Reading the record

- **WHEN** an administrator asks what has been done
- **THEN** they SHALL be shown it, newest first

#### Scenario: Reading it as anybody else

- **WHEN** somebody who is not an administrator asks
- **THEN** it SHALL be refused

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

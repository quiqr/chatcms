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

### Requirement: An administrator acts as somebody else, visibly

An administrator SHALL be able to act as another person, so that they see
and can do what that person sees and can do. It SHALL announce itself on the
screen while it is in force, in the administration record, and in the
history of whatever is changed.

#### Scenario: Beginning it

- **WHEN** an administrator begins acting as somebody
- **THEN** every screen SHALL answer as that person does
- **AND** what may be done SHALL be what that person may do, decided by the
  permission function rather than by a second set of rules

#### Scenario: Beginning it as anybody else

- **WHEN** somebody who does not administer the installation asks to act as
  another person
- **THEN** it SHALL be refused by the permission function, whatever they
  hold on any organisation

#### Scenario: Acting as an administrator

- **WHEN** an administrator asks to act as somebody who also administers the
  installation
- **THEN** it SHALL be refused, because that is a way of becoming one

#### Scenario: Acting as an account that is off

- **WHEN** an administrator asks to act as somebody whose account is off
- **THEN** it SHALL be refused

### Requirement: The session stays the administrator's

Acting as somebody SHALL NOT issue a session as that person. The
administrator's own session SHALL carry who is being acted as, so that who
is really there is never lost.

#### Scenario: Who is really there

- **WHEN** an administrator is acting as somebody
- **THEN** the answer to who is signed in SHALL say both: the person being
  acted as, and the administrator really there

#### Scenario: Ending it

- **WHEN** the administrator ends it
- **THEN** they SHALL be themselves again, with what they could do before

#### Scenario: Signing out

- **WHEN** the session ends, by signing out or by expiring
- **THEN** the acting SHALL end with it

### Requirement: What is done while acting names both

Every version, and every record, written while acting SHALL name the
administrator and the person acted as. The history a customer reads SHALL
never say that the customer did something an administrator did.

#### Scenario: A version

- **WHEN** something is saved while acting as somebody
- **THEN** its author SHALL read as the administrator acting as that person,
  both in ChatCMS and in git

#### Scenario: The record

- **WHEN** acting begins and when it ends
- **THEN** each SHALL be in the administration record, with who, as whom,
  and when

### Requirement: The platform admin has an area of its own

Administering the installation SHALL happen in an area that says what it is,
with its own shell, so that nobody is in it by accident.

#### Scenario: Where you are

- **WHEN** a platform admin is in the area
- **THEN** the shell SHALL say so, and SHALL offer the way back to their own
  projects

#### Scenario: Anybody else

- **WHEN** somebody who does not administer the installation asks for the
  area
- **THEN** it SHALL be refused, and nothing SHALL link to it from their
  screens

### Requirement: Every organisation is listed with what it holds

The list of organisations SHALL say, for each one, who administers it, what
it holds against its limits, and when it was made. Searching and filtering
SHALL be answered by the server.

#### Scenario: One with nobody to administer it

- **WHEN** an organisation has no admin
- **THEN** the list SHALL say so, in the colour the design gives a warning

#### Scenario: Filtering

- **WHEN** a platform admin asks for the organisations at a limit, or those
  with no admin
- **THEN** only those SHALL be listed, decided by the server

#### Scenario: Searching

- **WHEN** a platform admin searches by name or address
- **THEN** only the organisations that match SHALL be listed

### Requirement: Removing an organisation removes what it holds in ChatCMS

Removing an organisation SHALL need administration of the installation, and
SHALL remove its sites and its memberships from ChatCMS. It SHALL NOT touch
anything in git, and SHALL be refused until the organisation's name has been
typed.

This replaces the rule that removing was refused while an organisation still
held sites. The hand-off of 2 November draws the opposite: the dialog lists
what goes and asks for the name, which is the protection the refusal stood
in for.

#### Scenario: What goes

- **WHEN** an organisation is removed
- **THEN** its sites SHALL no longer be in ChatCMS
- **AND** the accounts of its members SHALL remain, having lost their access
  to it

#### Scenario: What stays

- **WHEN** an organisation is removed
- **THEN** every git remote its sites used SHALL be untouched
- **AND** every live website SHALL keep working

#### Scenario: Asked for twice

- **WHEN** the name typed does not match the organisation's
- **THEN** removing SHALL NOT be offered

#### Scenario: Who may

- **WHEN** somebody who does not administer the installation asks to remove
  an organisation
- **THEN** it SHALL be refused

#### Scenario: What is written down

- **WHEN** an organisation is removed
- **THEN** the act SHALL be in the administration record, with how many sites
  went with it

### Requirement: An account is invited, active or blocked

An account SHALL be in one of three states, and every screen that shows an
account SHALL say which. Invited means made and never accepted; blocked means
it may not be used; active is neither.

#### Scenario: Somebody who has not accepted

- **WHEN** an account has been made and no password has ever been set on it
- **THEN** it SHALL be shown as invited

#### Scenario: Blocking

- **WHEN** a platform admin blocks an account
- **THEN** it SHALL not be able to sign in, and its sessions SHALL stop
- **AND** it SHALL be shown as blocked until it is unblocked

#### Scenario: Your own

- **WHEN** a platform admin asks to block or delete their own account
- **THEN** it SHALL be refused

### Requirement: An account can be managed by whoever administers the installation

A platform admin SHALL be able to change an account's name and address, send
it a password reset, put it into an organisation, make it a platform admin or
stop it being one, block it, and delete it. Each SHALL be written into the
administration record.

#### Scenario: Changing a name and an address

- **WHEN** a platform admin changes an account's name or address
- **THEN** the account SHALL carry them
- **AND** an address that already belongs to another account SHALL be refused

#### Scenario: Sending a reset

- **WHEN** a platform admin sends a password reset
- **THEN** the same link the product sends SHALL go out
- **AND** no password SHALL be set by anybody but the person themselves

#### Scenario: Deleting an account

- **WHEN** a platform admin deletes an account
- **THEN** the account and its memberships SHALL be gone
- **AND** what that person wrote SHALL still read as theirs

#### Scenario: Anybody else

- **WHEN** somebody who does not administer the installation asks for any of
  these
- **THEN** it SHALL be refused

### Requirement: The list of accounts is asked for a page at a time

The list SHALL be filtered and paged by the server, and SHALL say how many
there are altogether.

#### Scenario: A filter

- **WHEN** a platform admin asks for the platform admins, the invited, the
  blocked, or those in no organisation
- **THEN** only those SHALL be listed

#### Scenario: The next page

- **WHEN** a platform admin asks for more
- **THEN** the accounts after the last one shown SHALL be listed
- **AND** an account made while they were reading SHALL NOT make a row appear
  twice or not at all

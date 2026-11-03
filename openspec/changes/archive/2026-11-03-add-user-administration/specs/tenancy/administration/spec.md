## ADDED Requirements

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

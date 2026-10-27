# identity/accounts Specification

## Purpose
Defines what a person may change about their own account, and what turning an
account off means.

## Requirements

### Requirement: An account can be turned off, and on again

An account SHALL be able to be turned off by somebody who administers the
installation, and turned on again. An account that is off SHALL NOT be able
to sign in, and its sessions SHALL stop at once.

An account belongs to the installation rather than to an organisation, so an
organisation's owner SHALL NOT turn one off: somebody who belongs to two
organisations would otherwise be locked out of one by the other.

#### Scenario: Turning one off

- **WHEN** an account is turned off
- **THEN** signing in with it SHALL be refused
- **AND** any session it already had SHALL stop working immediately

#### Scenario: Turning it on again

- **WHEN** it is turned on again
- **THEN** signing in SHALL work
- **AND** the password SHALL be the one it had

#### Scenario: What they wrote

- **WHEN** an account is turned off
- **THEN** every version they made SHALL still name them

#### Scenario: Turning off your own

- **WHEN** somebody asks to turn off their own account
- **THEN** it SHALL be refused, and the message SHALL say why

#### Scenario: An organisation's owner

- **WHEN** an organisation's owner asks to turn off an account
- **THEN** it SHALL be refused, because the account is not theirs to hold

### Requirement: A person manages their own account

A person SHALL be able to change their own name and their own password, and
SHALL NOT be able to change anybody else's.

#### Scenario: Changing a name

- **WHEN** somebody changes their own name
- **THEN** it SHALL be what every screen shows for them from then on

#### Scenario: Changing a password

- **WHEN** somebody changes their password
- **THEN** the one they have SHALL be asked for first
- **AND** a wrong one SHALL be refused without changing anything

#### Scenario: A new password that is too weak

- **WHEN** the new password is refused by the same rules as any other
- **THEN** nothing SHALL change, and the reason SHALL be said

#### Scenario: Somebody else's account

- **WHEN** somebody asks to change another person's name or password
- **THEN** it SHALL be refused

### Requirement: When somebody was last here is known

Each account SHALL carry when it was last seen, so that the people screen can
say it.

#### Scenario: Somebody who has signed in

- **WHEN** somebody has used the application
- **THEN** when they were last seen SHALL be reported

#### Scenario: Somebody who has never accepted

- **WHEN** somebody has been invited and never signed in
- **THEN** the answer SHALL say so rather than giving a time

### Requirement: A person can have a picture

A person SHALL be able to give their account a picture, replace it, and
remove it. The picture SHALL be shown wherever that person is shown.

#### Scenario: Adding one

- **WHEN** somebody uploads a picture
- **THEN** it SHALL be shown for them from then on

#### Scenario: Something that is not a picture

- **WHEN** the file is not a picture, whatever it is named
- **THEN** it SHALL be refused, and nothing SHALL be stored

#### Scenario: Something too large

- **WHEN** the file is larger than the installation allows
- **THEN** it SHALL be refused before it is read into memory

#### Scenario: Removing it

- **WHEN** somebody removes their picture
- **THEN** their initials SHALL be shown again

#### Scenario: Somebody else's picture

- **WHEN** somebody asks to change another person's picture
- **THEN** it SHALL be refused

#### Scenario: Never in a customer's site

- **WHEN** a site is published
- **THEN** no person's picture SHALL be among what is published

### Requirement: An email address changes only once it is proven

Changing the address an account signs in with SHALL need the new address
proven before it takes effect.

#### Scenario: Asking for a new address

- **WHEN** somebody asks to change their address
- **THEN** a message SHALL be sent to the new one
- **AND** signing in SHALL still use the old one

#### Scenario: Proving it

- **WHEN** the link in that message is followed
- **THEN** the account SHALL use the new address
- **AND** the old one SHALL stop working

#### Scenario: An address somebody else holds

- **WHEN** the new address already belongs to an account
- **THEN** it SHALL be refused

#### Scenario: A link that has run out

- **WHEN** the link has expired or been used
- **THEN** the address SHALL NOT change, and the screen SHALL say what to do

### Requirement: Somebody can see and end their sessions

A person SHALL be able to see every session their account holds and end any
one of them.

#### Scenario: Seeing them

- **WHEN** somebody asks where they are signed in
- **THEN** each session SHALL be listed with when it was last used
- **AND** the one they are reading on SHALL be marked

#### Scenario: Ending one

- **WHEN** somebody ends one session
- **THEN** that session SHALL stop working
- **AND** the others SHALL be untouched

#### Scenario: Somebody else's sessions

- **WHEN** somebody asks to end a session that is not theirs
- **THEN** it SHALL be refused

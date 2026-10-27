## Purpose

An account has a picture, an email address that can change, and sessions a
person can see and end.

## ADDED Requirements

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

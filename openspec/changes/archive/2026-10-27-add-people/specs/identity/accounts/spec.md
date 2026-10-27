## Purpose

Defines what a person may change about their own account, and what turning an
account off means.

## ADDED Requirements

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

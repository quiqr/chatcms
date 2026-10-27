## Purpose

Signing out ends a session, and can end every other one.

## ADDED Requirements

### Requirement: Somebody can sign out from anywhere in the application

Every screen SHALL offer a way to sign out and a way to reach your own
account.

#### Scenario: From a site's screens

- **WHEN** somebody is looking at any screen of a project
- **THEN** signing out SHALL be offered
- **AND** their own account SHALL be reachable

#### Scenario: On a phone

- **WHEN** the phone layout is shown
- **THEN** both SHALL be offered there too

#### Scenario: After signing out

- **WHEN** somebody signs out
- **THEN** the session SHALL be ended
- **AND** the old session value SHALL be refused from then on
- **AND** the screen SHALL say that they signed out, which a session that
  merely ran out SHALL NOT say

### Requirement: Somebody can sign out everywhere

Somebody SHALL be able to end every session their account has, not only the
one they are using.

#### Scenario: Ending them all

- **WHEN** somebody signs out everywhere
- **THEN** every session that account holds SHALL be ended
- **AND** the one that asked SHALL be ended with them

#### Scenario: Somebody else's sessions

- **WHEN** somebody signs out everywhere
- **THEN** no other account's sessions SHALL be touched

#### Scenario: Not signed in

- **WHEN** nobody is signed in and signing out everywhere is asked for
- **THEN** it SHALL be refused

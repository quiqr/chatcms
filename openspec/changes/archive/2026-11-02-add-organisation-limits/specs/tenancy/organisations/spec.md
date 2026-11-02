## ADDED Requirements

### Requirement: An organisation may have limits, and they are optional

An organisation MAY have a maximum number of members and a maximum number of
sites. Each is independent of the other, and an organisation with neither
SHALL be unlimited. Only a platform admin SHALL set either.

#### Scenario: A new organisation

- **WHEN** an organisation is made
- **THEN** it SHALL have no limits
- **AND** nothing about it SHALL be refused for being full

#### Scenario: Setting one

- **WHEN** a platform admin sets a limit
- **THEN** it SHALL hold from that moment
- **AND** somebody who administers the organisation SHALL NOT be able to
  change it

#### Scenario: Lowering it below what is in use

- **WHEN** a limit is set below what the organisation already has
- **THEN** nothing SHALL be removed
- **AND** nothing new SHALL be allowed until usage is under the limit again

### Requirement: A limit is refused before the act runs

Every act that would take an organisation over one of its limits SHALL be
refused by the server before it runs. The refusal SHALL name the limit, the
numbers, and what can be done about it.

#### Scenario: One more member than the limit allows

- **WHEN** somebody is put into an organisation whose seats are all held
- **THEN** it SHALL be refused
- **AND** the refusal SHALL say how many that would make and what the limit is

#### Scenario: One more site than the limit allows

- **WHEN** a site is made in an organisation whose sites are all in use
- **THEN** it SHALL be refused

#### Scenario: Somebody invited and not yet accepted

- **WHEN** the seats are counted
- **THEN** somebody who has been invited and has not accepted SHALL hold one

#### Scenario: A screen that would be refused

- **WHEN** a screen offers an act that would be refused for a limit
- **THEN** the act SHALL be offered disabled, with the reason beside it,
  rather than refused after it is pressed

### Requirement: How full an organisation is, is read in one place

What an organisation holds, and what it is allowed, SHALL be read through one
function, so that no two screens can disagree about how full it is.

#### Scenario: Two screens, one answer

- **WHEN** an organisation's own screens and the platform admin's screens
  show how full it is
- **THEN** both SHALL read the same numbers

#### Scenario: No limit

- **WHEN** an organisation has no limit for something
- **THEN** what it holds SHALL still be counted and shown, and nothing SHALL
  be shown about how much room is left

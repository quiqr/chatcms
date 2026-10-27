## Purpose

An organisation's own owner manages who belongs to it.

## ADDED Requirements

### Requirement: An organisation's owner invites people to it

Somebody who owns an organisation SHALL be able to invite an address to it,
change what that person holds there, and take them out.

#### Scenario: Inviting an address with no account

- **WHEN** an owner invites an address nobody holds
- **THEN** an account SHALL be made for it
- **AND** that person SHALL be sent the way in
- **AND** they SHALL belong to the organisation

#### Scenario: Inviting somebody who already has an account

- **WHEN** an owner invites an address that has one
- **THEN** that person SHALL belong to the organisation
- **AND** no way in SHALL be sent, because they have one

#### Scenario: Somebody who merely belongs

- **WHEN** a member who does not own the organisation asks to invite
- **THEN** it SHALL be refused

#### Scenario: Taking somebody out

- **WHEN** an owner takes somebody out
- **THEN** they SHALL no longer belong to the organisation

#### Scenario: The last owner

- **WHEN** taking somebody out or changing their role would leave the
  organisation with no owner
- **THEN** it SHALL be refused, and the message SHALL say why

### Requirement: An invitation that was never accepted can be sent again or withdrawn

An invitation SHALL be able to be sent again, and to be withdrawn.

#### Scenario: Sending it again

- **WHEN** an invitation that has not been accepted is sent again
- **THEN** a new way in SHALL be sent
- **AND** the previous one SHALL stop working

#### Scenario: Withdrawing it

- **WHEN** an invitation that has not been accepted is withdrawn
- **THEN** that person SHALL no longer belong to what they were invited to
- **AND** the way in they were sent SHALL stop working

#### Scenario: Somebody who has accepted

- **WHEN** sending again or withdrawing is asked for somebody who has
  accepted
- **THEN** it SHALL be refused, because there is no invitation to act on

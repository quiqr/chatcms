## ADDED Requirements

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

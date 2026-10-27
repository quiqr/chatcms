# tenancy/organisations Specification

## Purpose
Defines what an organisation and a project are, who belongs to them, and how
someone is brought in, so that work has an owner and a boundary.

## Requirements

### Requirement: An organisation holds projects and people

An organisation SHALL have a name and an address of its own, SHALL hold
projects, and SHALL have at least one owner at all times. Only an
administrator of the installation SHALL create one.

#### Scenario: Creating an organisation

- **WHEN** an administrator creates an organisation and names its owner
- **THEN** it exists with that person as its owner

#### Scenario: Creating one without being an administrator

- **WHEN** somebody who does not administer the installation asks to create an
  organisation
- **THEN** it is refused

#### Scenario: The last owner cannot be removed

- **WHEN** removing a person would leave an organisation with no owner
- **THEN** the removal is refused, and the message says why

#### Scenario: Two organisations may share a person

- **WHEN** a person belongs to two organisations
- **THEN** each is separate, and what they may do in one says nothing about
  the other

### Requirement: A project belongs to one organisation

A project SHALL belong to exactly one organisation, SHALL have a name and an
address of its own within it, and SHALL be reachable only by people with a
role on it or ownership of its organisation.

#### Scenario: Creating a project

- **WHEN** an organisation's owner creates a project
- **THEN** it exists in that organisation

#### Scenario: Somebody else's project

- **WHEN** someone asks for a project they may not reach
- **THEN** the answer does not distinguish a project they may not reach from
  one that does not exist

#### Scenario: Listing projects

- **WHEN** someone lists the projects of an organisation
- **THEN** they see the ones they may reach and no others

### Requirement: An invitation gives a role and a way in

Inviting someone SHALL take an address and a role. A person who does not exist
yet SHALL be created with no password and given a link that sets one; a person
who already exists SHALL be given the role.

#### Scenario: Inviting someone new

- **WHEN** an owner invites an address nobody has, as an editor
- **THEN** a person exists with that address and no password
- **AND** they hold the editor role on the project
- **AND** a link that sets their first password has been issued

#### Scenario: Inviting someone already known

- **WHEN** an owner invites an address that already belongs to a person
- **THEN** that person is given the role
- **AND** no second person is made

#### Scenario: Inviting someone who already has a role

- **WHEN** an owner invites someone who already holds a role on the project
- **THEN** their role is changed to the one invited with, rather than doubled

#### Scenario: Only those who may, may

- **WHEN** an editor or a viewer invites someone
- **THEN** they are refused

### Requirement: A role can be changed and taken away

Someone who may manage people SHALL be able to change another person's role
and remove them from a project.

#### Scenario: Changing a role

- **WHEN** an owner changes an editor to a viewer
- **THEN** that person may no longer change anything

#### Scenario: Removing a person

- **WHEN** an owner removes a person from a project
- **THEN** that person can no longer reach it

#### Scenario: Removing oneself as the last owner

- **WHEN** the only person who can administer a project would be removed
- **THEN** the removal is refused

### Requirement: What someone may do is told to the interface

A person's own view SHALL carry what they may do, so that the interface leaves
out what they may not rather than showing it disabled.

#### Scenario: Asking what this person may do

- **WHEN** someone asks for a project they can reach
- **THEN** the answer says which of the things they might do they may do

#### Scenario: An editor's view

- **WHEN** an editor asks
- **THEN** the answer says they may not manage people and may not see
  technical detail

### Requirement: A project records where it came from

A project SHALL record how its site arrived, so that an implementer can see
later what was imported and from where.

#### Scenario: Reading a project

- **WHEN** a project that was imported is read
- **THEN** it says which kind of source it came from and, where there is one,
  its address

#### Scenario: Nothing imported yet

- **WHEN** a project has had nothing imported
- **THEN** it says so rather than showing an empty address

### Requirement: An organisation can be renamed

An organisation's name SHALL be changeable by an administrator of the
installation, and the new name SHALL be what every screen shows from then on.
Its address SHALL NOT change with it, because links already written would
stop working.

#### Scenario: Renaming it

- **WHEN** an administrator gives an organisation a new name
- **THEN** it SHALL be called that
- **AND** its address SHALL be what it was

#### Scenario: Renaming it as its owner

- **WHEN** the organisation's own owner asks to rename it
- **THEN** it SHALL be refused

#### Scenario: A name with nothing in it

- **WHEN** the new name is empty or only spaces
- **THEN** it SHALL be refused, and the old name SHALL stand

### Requirement: The people of an organisation can be seen and changed

Somebody who owns an organisation SHALL see who is in it and what each holds,
and SHALL be able to change a role or take somebody out.

#### Scenario: Its owner

- **WHEN** an organisation's owner asks who is in it
- **THEN** every person SHALL be listed with the role they hold there

#### Scenario: Somebody who is merely in it

- **WHEN** a member who does not own it asks to change a role
- **THEN** it SHALL be refused

#### Scenario: Somebody with no part in it

- **WHEN** somebody who does not belong asks who is in it
- **THEN** the answer SHALL be that there is nothing there

#### Scenario: The last owner

- **WHEN** a change would leave an organisation with no owner
- **THEN** it SHALL be refused, and the message SHALL say why

### Requirement: A screen knows which organisation it is showing

Every screen SHALL take the organisation it is showing from the address, and
SHALL NOT assume one.

#### Scenario: An address naming an organisation somebody is not in

- **WHEN** somebody opens an address for an organisation they do not belong
  to and do not administer
- **THEN** the answer SHALL be that there is nothing there
- **AND** they SHALL NOT be sent to one they do belong to

#### Scenario: Two organisations

- **WHEN** somebody belongs to two and switches between them
- **THEN** each SHALL show its own projects
- **AND** nothing from one SHALL appear in the other

#### Scenario: Belonging to none

- **WHEN** somebody who belongs to no organisation opens their projects
- **THEN** they SHALL be told so, and told what to do about it
- **AND** an administrator SHALL additionally be offered to start one

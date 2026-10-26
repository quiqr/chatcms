# tenancy/permissions Specification

## Purpose
Defines who may do what, and establishes that one function decides it, so that
a permission can be changed in one place and cannot be forgotten in another.

## Requirements

### Requirement: One function decides every permission

Every permission decision SHALL be made by a single function taking who is
asking, what they want to do, and what they want to do it to. No other part of
the system SHALL compare a role.

#### Scenario: The rule is enforced, not documented

- **WHEN** any package other than the one holding that function mentions a
  role
- **THEN** the test suite fails and names that package

#### Scenario: A handler asks rather than decides

- **WHEN** an endpoint needs to know whether it may proceed
- **THEN** it calls that function and acts on the answer

### Requirement: The decision takes a resource

The function SHALL take the thing being acted on, even where today's answer
does not depend on it, so that a finer permission can be added by changing one
function.

#### Scenario: The resource is part of the question

- **WHEN** a permission is asked about
- **THEN** the question names the organisation or project concerned, and may
  name something within it

#### Scenario: A resource outside the actor's reach

- **WHEN** someone asks about a project they hold no role on, and whose
  organisation they do not own
- **THEN** the answer is no, whatever the action

### Requirement: The four roles have the powers the designs give them

The roles SHALL be owner, implementer, editor and viewer, with exactly the
powers set out in the design hand-off.

#### Scenario: Editing content and media

- **WHEN** an owner, an implementer or an editor acts on content or media
- **THEN** they may change it
- **AND** a viewer may read it and not change it

#### Scenario: Publishing and going back to a version

- **WHEN** an owner, an implementer or an editor publishes, or goes back to an
  earlier version
- **THEN** they may
- **AND** a viewer may not

#### Scenario: Setting up a project

- **WHEN** an owner or an implementer manages people, project settings or
  workflows
- **THEN** they may
- **AND** an editor and a viewer may not

#### Scenario: Seeing technical detail

- **WHEN** an owner or an implementer asks for technical detail
- **THEN** they may see it
- **AND** an editor and a viewer may not

#### Scenario: Billing and deleting a project

- **WHEN** an owner bills or deletes a project
- **THEN** they may
- **AND** an implementer, an editor and a viewer may not

#### Scenario: The whole table is checked

- **WHEN** the permissions are tested
- **THEN** every combination of role and action is checked against an answer
  written out rather than worked out

### Requirement: Owning an organisation owns its projects

Someone who owns an organisation SHALL hold the owner role on every project in
it, without needing a role on each.

#### Scenario: An organisation owner reaches a project

- **WHEN** an organisation's owner acts on a project in it that they hold no
  role on
- **THEN** they are treated as its owner

#### Scenario: An organisation member does not

- **WHEN** an organisation's member acts on a project in it that they hold no
  role on
- **THEN** they are treated as holding no role

### Requirement: Nobody is somebody with no powers

An actor who is not signed in SHALL be refused everything.

#### Scenario: Nobody asks

- **WHEN** a permission is asked about for nobody
- **THEN** the answer is no, whatever the action and whatever the resource

### Requirement: The decision knows the installation as well as the resource

The permission function SHALL answer for actions held over the installation
itself, not only for actions on an organisation or a project, and SHALL do so
without granting any permission over a resource.

#### Scenario: An action over the installation

- **WHEN** the function is asked about an action that needs administration
- **THEN** it SHALL answer from whether the actor administers the installation

#### Scenario: An action on a project

- **WHEN** the function is asked about an action on a project
- **THEN** administering the installation SHALL NOT change the answer

#### Scenario: The rule still holds

- **WHEN** any package other than the one holding that function compares
  whether somebody is an administrator
- **THEN** the test suite SHALL fail and name that package

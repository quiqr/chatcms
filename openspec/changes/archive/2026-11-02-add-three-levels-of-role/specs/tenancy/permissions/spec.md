## RENAMED Requirements

- FROM: `### Requirement: The four roles have the powers the designs give them`
- TO: `### Requirement: The three roles have the powers the designs give them`
## MODIFIED Requirements

### Requirement: The three roles have the powers the designs give them

The project roles SHALL be implementer, editor and viewer, with exactly the
powers set out in the design hand-off. There SHALL be no owner role on a
project; what it held belongs to whoever administers the organisation.

#### Scenario: Editing content and media

- **WHEN** an implementer or an editor acts on content or media
- **THEN** they may change it
- **AND** a viewer may read it and not change it

#### Scenario: Publishing and going back to a version

- **WHEN** an implementer or an editor publishes, or goes back to an earlier
  version
- **THEN** they may
- **AND** a viewer may not

#### Scenario: Setting up a project

- **WHEN** an implementer manages people, project settings or workflows
- **THEN** they may
- **AND** an editor and a viewer may not

#### Scenario: Seeing technical detail

- **WHEN** an implementer asks for technical detail
- **THEN** they may see it
- **AND** an editor and a viewer may not

#### Scenario: Billing and deleting a project

- **WHEN** somebody who administers the organisation bills or deletes a
  project in it
- **THEN** they may
- **AND** an implementer, an editor and a viewer may not, whatever role they
  hold on the project

#### Scenario: The whole table is checked

- **WHEN** the permissions are tested
- **THEN** every combination of role and action is checked against an answer
  written out rather than worked out

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
  organisation they do not administer
- **THEN** the answer is no, whatever the action

## ADDED Requirements

### Requirement: Administering an organisation reaches its projects

Somebody who administers an organisation SHALL hold the implementer role on
every project in it, without needing a role on each. They SHALL NOT thereby
hold a role nobody has: there is no owner of a project.

#### Scenario: An organisation admin reaches a project

- **WHEN** somebody who administers an organisation acts on a project in it
  that they hold no role on
- **THEN** they are treated as its implementer

#### Scenario: An organisation member does not

- **WHEN** an organisation's member acts on a project in it that they hold no
  role on
- **THEN** they are treated as holding no role

#### Scenario: A role belongs to a place

- **WHEN** somebody administers one organisation and holds a project role in
  another
- **THEN** each answer is decided from where the resource is, and neither
  reaches into the other


### Requirement: Three levels, each managing the one below

The installation SHALL have three levels of role: a platform admin over the
whole installation, an organisation admin over one organisation, and project
roles within one project. Each level SHALL manage the one below it, and no
level SHALL be reachable from below.

#### Scenario: Moving a project to another organisation

- **WHEN** a platform admin moves a project
- **THEN** they may move it to any organisation
- **AND** an organisation admin may move it only to an organisation they also
  administer
- **AND** nobody else may move it at all

#### Scenario: Setting an organisation's limits

- **WHEN** somebody asks to set the limits on an organisation
- **THEN** only a platform admin may

#### Scenario: What a platform admin holds over content

- **WHEN** a platform admin acts on a customer's content
- **THEN** administering the installation SHALL NOT be what allows it

## REMOVED Requirements

### Requirement: Owning an organisation owns its projects

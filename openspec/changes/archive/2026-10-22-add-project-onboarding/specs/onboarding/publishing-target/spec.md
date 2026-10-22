## Purpose

Defines where a site goes when it is published and in what form, so that
ChatCMS fits the pipeline a customer already has rather than replacing it.

## ADDED Requirements

### Requirement: A project records where it publishes

A project SHALL record whether it publishes to the remote's main line or to a
line of its own, and the name of that line when it is its own.

#### Scenario: Publishing to the main line

- **WHEN** a project is set to publish to the main line
- **THEN** publishing sends work there

#### Scenario: Publishing to its own line

- **WHEN** a project is set to publish to a line of its own
- **THEN** that line's name is recorded, with a sensible default offered

#### Scenario: A name that cannot be used

- **WHEN** a name is given that the remote cannot accept
- **THEN** it is refused with a message saying what is allowed

### Requirement: A project records what it publishes

A project SHALL record whether it publishes the site's source, for a pipeline
elsewhere to build, or the result of building it here.

#### Scenario: Publishing the source

- **WHEN** a project publishes its source
- **THEN** what is sent is the site's files as they are

#### Scenario: Publishing the built result

- **WHEN** a project publishes the built result
- **THEN** what is sent is what building the site produced

#### Scenario: The choice is remembered and shown

- **WHEN** the choice has been made
- **THEN** it is shown as the current setting rather than asked again

### Requirement: The settings are only for those who may

Changing where or what a project publishes SHALL be allowed only to those who
may change the project's settings.

#### Scenario: An implementer changes them

- **WHEN** an owner or an implementer changes a publishing setting
- **THEN** it is changed

#### Scenario: An editor cannot

- **WHEN** an editor or a viewer changes a publishing setting
- **THEN** they are refused and nothing changes

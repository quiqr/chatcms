## Purpose

Names who may bring an organisation into being, which was left unsaid.

## MODIFIED Requirements

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

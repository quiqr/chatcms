## Purpose

Adds the level above an organisation to the one function that decides
permissions.

## ADDED Requirements

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

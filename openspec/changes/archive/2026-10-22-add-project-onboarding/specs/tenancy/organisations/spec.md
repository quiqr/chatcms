## ADDED Requirements

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

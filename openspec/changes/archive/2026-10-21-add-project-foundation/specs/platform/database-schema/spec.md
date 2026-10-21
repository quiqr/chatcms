## Purpose

Defines how the ChatCMS database schema is created and evolved, and the
guarantee that a deployment never serves traffic against a schema it does not
understand.

## ADDED Requirements

### Requirement: Migrations ship inside the binary

The schema SHALL be defined by ordered migrations embedded in the server
binary. Running the server SHALL require no migration files on disk and no
external migration tool.

#### Scenario: The binary is run from an empty directory

- **WHEN** the server binary is run from a directory containing nothing
- **THEN** it can still migrate a database to the current schema version

#### Scenario: The binary reports its schema version

- **WHEN** the health endpoint is requested after a successful migration
- **THEN** it reports the version of the last applied migration

### Requirement: Migrations run before the listener opens

The server SHALL apply all pending migrations during start-up and SHALL open
its listener only after they succeed. A migration failure SHALL be fatal.

#### Scenario: An empty database

- **WHEN** the server starts against a database with no tables
- **THEN** every migration is applied in order
- **AND** the listener opens afterwards

#### Scenario: An already migrated database

- **WHEN** the server starts against a database already at the current version
- **THEN** no migration is applied
- **AND** start-up is not slower in any way the operator must plan for

#### Scenario: A migration fails

- **WHEN** a migration fails to apply
- **THEN** the process exits non-zero
- **AND** no listener was ever opened
- **AND** the log names the migration that failed

#### Scenario: A database newer than the binary

- **WHEN** the database has migrations applied that this binary does not know
- **THEN** the process exits non-zero with a message saying the binary is older
  than the database
- **AND** no listener is opened

### Requirement: Two servers may start at once

Migrations SHALL be safe to run concurrently. When two instances start
together, exactly one SHALL apply a given migration and both SHALL end up
serving.

#### Scenario: Concurrent start-up

- **WHEN** two server processes start against the same unmigrated database
- **THEN** each migration is applied exactly once
- **AND** both processes open their listeners
- **AND** neither reports a migration failure

### Requirement: Query code is generated from checked-in SQL

Database queries SHALL be written as SQL and the Go access code generated from
them. The generated code SHALL be committed, so building the server requires no
code generator.

#### Scenario: Building without the generator

- **WHEN** the server is built in an environment without the query generator
- **THEN** the build succeeds

#### Scenario: Generated code is stale

- **WHEN** the committed generated code does not match the SQL it came from
- **THEN** the gate fails and names the file that is out of date

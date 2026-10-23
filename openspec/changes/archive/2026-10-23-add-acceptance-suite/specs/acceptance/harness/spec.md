## Purpose

Defines what a run of the acceptance suite owns, what it borrows, and what it
is never allowed to do.

## ADDED Requirements

### Requirement: A run owns everything it touches

A run SHALL create its own PostgreSQL, its own bare remote, its own seeded
site and its own server, and SHALL remove them afterwards.

#### Scenario: two specs at once

- **WHEN** one spec changes a page and another publishes
- **THEN** neither SHALL see the other's database, remote or working copy

#### Scenario: a spec fails

- **WHEN** a spec fails part way through
- **THEN** its database, server and remote SHALL still be cleaned up

#### Scenario: nothing is left behind

- **WHEN** the suite finishes
- **THEN** no PostgreSQL, server or working copy from it SHALL still be
  running

#### Scenario: a test binary that starts a database

- **WHEN** a test binary starts a PostgreSQL of its own
- **THEN** that PostgreSQL SHALL stop when the binary ends, however it ends

### Requirement: A run reaches nothing outside itself

A run SHALL NOT need the network, and SHALL use only programs pinned by the
flake.

#### Scenario: in the build sandbox

- **WHEN** the suite runs as the `e2e` check
- **THEN** it SHALL pass with no network at all

#### Scenario: the programs it uses

- **WHEN** the suite starts a server, a site build or a preview
- **THEN** the git, Hugo and browser used SHALL be the ones this flake pins

### Requirement: The suite is a check in the gate

`nix flake check` SHALL run the acceptance suite, so that a change that
breaks one of the nine cannot be shipped.

#### Scenario: a broken acceptance test

- **WHEN** a change breaks an acceptance test
- **THEN** `nix flake check` SHALL fail

### Requirement: The same specs can be run against a real site

A run SHALL be able to use a real repository in place of the fixture site,
without the specs being changed.

#### Scenario: pointed at a repository

- **WHEN** the suite is pointed at a repository and a model within it
- **THEN** the same specs SHALL run against that site

#### Scenario: a site the specs cannot assume

- **WHEN** a real site has no page a spec names
- **THEN** that spec SHALL say so plainly rather than fail as though the
  product were broken

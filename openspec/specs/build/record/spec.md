# build/record Specification

## Purpose
Defines what is remembered about every build, so that a failure can be read
afterwards by the person who has to fix it.

## Requirements

### Requirement: Every build is recorded

Each build SHALL be recorded with the project it was for, why it ran, when it
started, when it finished, whether it succeeded, and its log.

#### Scenario: a build that worked is recorded as such

- **WHEN** a build succeeds
- **THEN** a record SHALL exist carrying the project, the start, the finish and
  a state saying it succeeded

#### Scenario: a build that failed keeps its log

- **WHEN** a build fails
- **THEN** its record SHALL carry the log
- **AND** the log SHALL be readable by somebody who may see the project

#### Scenario: a build that was still running when the server stopped

- **WHEN** the server stops while a build is recorded as running
- **THEN** on the next start that record SHALL NOT stay running forever
- **AND** SHALL be reported as interrupted

### Requirement: A log is bounded

A build log SHALL be kept up to a stated size. When a build produces more, what
is kept SHALL say that it was shortened and SHALL keep the end, where the error
is.

#### Scenario: a very long log is shortened rather than refused

- **WHEN** a build writes more log output than the limit
- **THEN** the record SHALL still be written
- **AND** the kept log SHALL say that output was dropped
- **AND** SHALL contain the last part of the output

### Requirement: Build history is per project and recent first

The builds of a project SHALL be listable, most recent first, carrying no other
project's builds.

#### Scenario: listing builds

- **WHEN** the builds of a project are listed
- **THEN** they SHALL be ordered most recent first
- **AND** SHALL carry only that project's builds

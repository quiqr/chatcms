## Purpose

Defines how a preview is reached and who may see one.

## ADDED Requirements

### Requirement: A preview is reached through the app

A preview SHALL be served under the app's own address, and the preview server
itself SHALL listen only on the machine it runs on.

#### Scenario: somebody not signed in

- **WHEN** a preview is asked for by somebody not signed in
- **THEN** it SHALL be refused
- **AND** no part of the site SHALL be served

#### Scenario: the path within the site

- **WHEN** a preview is asked for a path inside the site
- **THEN** the preview SHALL be asked for that path, without the part of the
  address that names the project

#### Scenario: the front page

- **WHEN** a preview is asked for with no path at all
- **THEN** the site's front page SHALL be served

#### Scenario: an address naming no project

- **WHEN** the address names no project
- **THEN** the answer SHALL be that there is nothing there

### Requirement: Anybody who may read the project may preview it

Seeing a preview SHALL need the right to read the project and nothing more.

#### Scenario: a viewer

- **WHEN** somebody who may read but not change the site asks for a preview
- **THEN** they SHALL be shown it

#### Scenario: somebody with no part in the project

- **WHEN** somebody with no part in the project asks
- **THEN** they SHALL be told there is nothing there

### Requirement: A preview shows the site as a reader will see it

A preview SHALL serve the site built by the same pinned Hugo that builds it
for publishing.

#### Scenario: the site renders

- **WHEN** a preview of a site that builds is asked for
- **THEN** the site's own pages SHALL be served

#### Scenario: a site that will not build

- **WHEN** a preview of a site Hugo cannot build is asked for
- **THEN** it SHALL be refused with a reason worth showing
- **AND** no preview SHALL be left running
